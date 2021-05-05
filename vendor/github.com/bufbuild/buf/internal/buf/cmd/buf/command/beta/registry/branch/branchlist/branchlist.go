// Copyright 2020-2021 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package branchlist

import (
	"context"
	"fmt"

	"github.com/bufbuild/buf/internal/buf/bufcli"
	"github.com/bufbuild/buf/internal/buf/bufcore/bufmodule"
	"github.com/bufbuild/buf/internal/buf/bufprint"
	"github.com/bufbuild/buf/internal/pkg/app/appcmd"
	"github.com/bufbuild/buf/internal/pkg/app/appflag"
	"github.com/bufbuild/buf/internal/pkg/rpc"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	pageSizeFlagName  = "page-size"
	pageTokenFlagName = "page-token"
	reverseFlagName   = "reverse"
	formatFlagName    = "format"
)

// NewCommand returns a new Command
func NewCommand(
	name string,
	builder appflag.Builder,
) *appcmd.Command {
	flags := newFlags()
	return &appcmd.Command{
		Use:   name + " <buf.build/owner/repository>",
		Short: "List branches for the specified repository.",
		Args:  cobra.ExactArgs(1),
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appflag.Container) error {
				return run(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type flags struct {
	PageSize  uint32
	PageToken string
	Reverse   bool
	Format    string
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	flagSet.Uint32Var(&f.PageSize,
		pageSizeFlagName,
		10,
		`The page size.`,
	)
	flagSet.StringVar(&f.PageToken,
		pageTokenFlagName,
		"",
		`The page token.`,
	)
	flagSet.BoolVar(&f.Reverse,
		reverseFlagName,
		false,
		`Reverse the results.`,
	)
	flagSet.StringVar(
		&f.Format,
		formatFlagName,
		bufprint.FormatText.String(),
		fmt.Sprintf(`The output format to use. Must be one of %s`, bufprint.AllFormatsString),
	)
}

func run(
	ctx context.Context,
	container appflag.Container,
	flags *flags,
) error {
	if container.Arg(0) == "" {
		return appcmd.NewInvalidArgumentError("repository is required")
	}
	moduleIdentity, err := bufmodule.ModuleIdentityForString(container.Arg(0))
	if err != nil {
		return appcmd.NewInvalidArgumentError(err.Error())
	}
	apiProvider, err := bufcli.NewRegistryProvider(ctx, container)
	if err != nil {
		return err
	}
	repositoryService, err := apiProvider.NewRepositoryService(ctx, moduleIdentity.Remote())
	if err != nil {
		return err
	}
	ctx, err = bufcli.WithHeaders(ctx, container, moduleIdentity.Remote())
	if err != nil {
		return err
	}
	repository, err := repositoryService.GetRepositoryByFullName(ctx, moduleIdentity.Owner()+"/"+moduleIdentity.Repository())
	if err != nil {
		if rpc.GetErrorCode(err) == rpc.ErrorCodeNotFound {
			return bufcli.NewRepositoryNotFoundError(container.Arg(0))
		}
		return bufcli.NewRPCError("list branches", moduleIdentity.Remote(), err)
	}
	repositoryBranchService, err := apiProvider.NewRepositoryBranchService(ctx, moduleIdentity.Remote())
	if err != nil {
		return err
	}
	repositoryBranches, _, err := repositoryBranchService.ListRepositoryBranches(
		ctx,
		repository.Id,
		flags.PageSize,
		flags.PageToken,
		flags.Reverse,
	)
	if err != nil {
		return bufcli.NewRPCError("list branches", moduleIdentity.Remote(), err)
	}
	return bufcli.PrintRepositoryBranches(ctx, container.Stdout(), flags.Format, repositoryBranches...)
}
