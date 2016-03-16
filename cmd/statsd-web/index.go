package main

// TODO: package up chart.js?

var indexTmpl = `
<!doctype html>

<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Statsd-Web</title>
</head>

<body>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/1.0.2/Chart.min.js"></script>
  <script type="text/javascript">` + scriptContents + `</script>
  <div id="counters">
    <canvas id="Counters.counter-0" width="1024px" height="400"></canvas>
  </div>
</body>
</html>

`

var scriptContents = `
var lastIndex = -1;
var max = 15;

var charts = {}

function newChart(id) {
  var el = document.getElementById(id);
  if (!el) {
    console.log("No element found for id", id);
    return null;
  }

  var ctx =el.getContext("2d");
  return new Chart(ctx);
}

function addToChart(section, time, name, value) {
  var id = section + "." + name;

  var c = charts[id];
  if (!c) {
    c = newChart(id);
    if (!c) {
      return;
    }

    data = {
      labels: [],
      datasets: [
        {
          label: name,
          fillColor: "rgba(220,220,220,0.2)",
          strokeColor: "rgba(220,220,220,1)",
          pointColor: "rgba(220,220,220,1)",
          pointStrokeColor: "#fff",
          pointHighlightFill: "#fff",
          pointHighlightStroke: "rgba(220,220,220,1)",
        }
      ]
    };
    c = c.Line(data, {
      animation: false,
      bezierCurve: false
    });
    charts[id] = c;
  }

  c.addData([value], time);
  while (c.datasets[0].points.length > max) {
    c.removeData();
  }
}

function addSnapshot(ss) {
  for (var metric in ss.counters) {
      if (!ss.counters.hasOwnProperty(metric)) {
          continue;
      }

      addToChart("Counters", ss.key, metric, ss.counters[metric])
      console.log("add key", ss.key, metric, ss.counters[metric]);
  }
}

function processSnapshot(ss) {
  if (!ss || ss.length == 0) {
    return;
  }

  curLast = lastIndex;
  lastIndex = ss[ss.length-1].index;

  if (curLast >= lastIndex) {
    // TODO: Clear all the data.
    alert('data was reset');
  }

  for(var i = 0; i < ss.length; i++) {
    addSnapshot(ss[i]);
  }
}

function refreshData() {
  var req = new XMLHttpRequest();
  req.open("GET", "/json?max=100&from=" + (lastIndex+1), true /* async */);
  req.onload = function() {
    setTimeout(refreshData, 300);
    if (req.status != 200) {
      console.log("Request failed", req);
      return
    }

    processSnapshot(JSON.parse(req.responseText));
  }
  req.send();
}

setTimeout(refreshData, 500);

`
