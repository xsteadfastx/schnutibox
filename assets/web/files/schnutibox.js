if(typeof(EventSource) !== "undefined") {
        var logSource = new EventSource("/log");
        logSource.onmessage = function(event) {
                var j = JSON.parse(event.data);
                /* eslint-disable no-prototype-builtins */
                if (j.hasOwnProperty("message")) {
                        document.getElementById("log").innerHTML += j.message + "<br>";
                }
        };
} else {
        document.getElementById("log").innerHTML = "Sorry, your browser does not support server-sent events...";
}

if(typeof(EventSource) !== "undefined") {
  var currentsongSource = new EventSource("/currentsong");
  currentsongSource.onmessage = function(event){
    document.getElementById("currentsong").innerHTML = event.data
  };
} else {
        document.getElementById("currentsong").innerHTML = "Sorry, your browser does not support server-sent events...";
};

function handleSubmit(event, url) {
  event.preventDefault()

  var data = new FormData(event.target)
  var value = Object.fromEntries(data.entries())
  var jsonValue = JSON.stringify(value)

  console.log(jsonValue)

  var xhr = new XMLHttpRequest()
  xhr.open("POST", url)
  xhr.setRequestHeader("Content-Type", "application/json")
  xhr.send(jsonValue)
}

var timerForm = document.querySelector('#timerForm')
timerForm.addEventListener('submit', function(){handleSubmit(event, "/api/v1/timer")})
