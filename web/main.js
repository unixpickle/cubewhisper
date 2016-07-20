(function() {

  function initialize() {
    window.app.createWorker(function(err, worker) {
      if (err) {
        alert('Error loading RNN file: ' + err);
      } else {
        stream = new window.app.SampleStream();
        setupApp(worker, stream);
      }
    });
  }

  function setupApp(worker, stream) {
    stream.on('error', function(err) {
      alert('Error: ' + err);
    });
    stream.on('start', worker.start.bind(worker));
    stream.on('samples', worker.samples.bind(worker));

    worker.on('moves', function(m) {
      var label = document.getElementById('classification');
      label.textContent = m;
    });

    var runButton = document.getElementById('run-button');
    runButton.addEventListener('click', function() {
      if (runButton.textContent === 'Run') {
        runButton.textContent = 'Done';
        stream.start();
      } else {
        runButton.textContent = 'Run';
        stream.stop();
        worker.end();
      }
    });
  }

  window.addEventListener('load', initialize);

})();
