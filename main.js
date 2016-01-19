(function() {

  var stream;
  var trainButton;
  var signatures = {};

  function initialize() {
    stream = new window.app.WordStream();
    stream.on('error', function(err) {
      alert('Error: ' + err);
    });
    stream.start();

    trainButton = document.getElementById('start-training');
    trainButton.addEventListener('click', train);
  }

  function train() {
    trainButton.disabled = true;

    var words = ['R', 'F', 'B', 'L', 'D', 'U', 'Prime', 'Squared'];
    var index = 0;

    var trainingText = document.getElementById('training-status');
    var trainNext = function() {
      if (index === words.length) {
        trainingText.innerText = 'Done training';
        doneTraining();
        return;
      }
      var word = words[index];
      trainingText.innerText = 'Please say ' + word;
      stream.once('word', function(sig) {
        signatures[word] = sig;
        trainNext();
      });
      ++index;
    };

    trainNext();
  }

  function doneTraining() {
    var transcription = document.createElement('p');
    transcription.id = 'transcription';
    document.body.appendChild(transcription);

    stream.on('word', function(sig) {
      var bestGuessDiff = Infinity;
      var bestGuess = 'None';

      var keys = Object.keys(signatures);
      for (var i = 0, len = keys.length; i < len; ++i) {
        var word = keys[i];
        var wordSignature = signatures[word];
        var diff = sig.difference(wordSignature);
        if (diff < bestGuessDiff) {
          bestGuessDiff = diff;
          bestGuess = word;
        }
      }

      transcription.innerHTML += bestGuess + '&nbsp;';
    });
  }

  window.addEventListener('load', initialize);

})();
