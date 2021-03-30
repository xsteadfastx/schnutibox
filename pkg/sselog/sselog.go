// package sselog is work in progress to implement a writer that sends its logs
// to a http server side event.
package sselog

type SSELog struct {
	LogChan   chan []byte
	Receivers []chan string
}

func (l SSELog) Write(p []byte) (n int, err error) {
	wCount := 0
	// Send log message to all receiver channels.
	for _, i := range l.Receivers {
		i <- string(p)

		wCount = +len(p)
	}

	return wCount, nil
}
