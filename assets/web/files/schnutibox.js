if(typeof(EventSource) !== "undefined") {
        var source = new EventSource("/log");
        source.onmessage = function(event) {
                var j = JSON.parse(event.data);
                if (j.hasOwnProperty('message')) {
                        document.getElementById("log").innerHTML += j.message + "<br>";
                }
        };
} else {
        document.getElementById("log").innerHTML = "Sorry, your browser does not support server-sent events...";
}

$("#timerForm").submit(function(event){
  event.preventDefault();
  var $form = $(this),
    duration = $form.find("input[name='s']").val(),
    url = $form.attr("action");

  $.ajax({
    url: url,
    type: "POST",
    data: JSON.stringify({"duration": duration}),
    contentType: "application/json",
    dataType: "json"
  });

})
