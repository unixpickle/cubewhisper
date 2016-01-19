(function() {

  var STATE_NOT_RUNNING = 0;
  var STATE_REQUESTED = 1;
  var STATE_REQUEST_CANCELLED = 2;
  var STATE_RUNNING = 3;

  var SAMPLE_INTERVAL = 20;
  var AMPLITUDE_BUFFER_SIZE = Math.ceil(400 / SAMPLE_INTERVAL);
  var AMPLITUDE_STOP_COUNT = Math.ceil(200 / SAMPLE_INTERVAL);
  var FFT_BUFFER_SIZE = Math.ceil(2000 / SAMPLE_INTERVAL);

  function WordStream() {
    window.EventEmitter.call(this);
    this._state = STATE_NOT_RUNNING;
    this._stream = null;
    this._analyser = null;
    this._interval = null;

    this._previousAmplitudes = [];
    this._previousFFTs = [];
  }

  WordStream.prototype = Object.create(window.EventEmitter.prototype);
  WordStream.prototype.constructor = WordStream;

  WordStream.prototype.start = function() {
    switch (this._state) {
    case STATE_NOT_RUNNING:
      this._requestUserMedia();
      break;
    case STATE_REQUEST_CANCELLED:
      this._state = STATE_REQUESTED;
      break;
    }
  };

  WordStream.prototype.stop = function() {
    switch (this._state) {
    case STATE_RUNNING:
      this._stopRunning();
      break;
    case STATE_REQUESTED:
      this._state = STATE_REQUEST_CANCELLED;
      break;
    }
  };

  WordStream.prototype._requestUserMedia = function() {
    this._state = STATE_REQUESTED;
    getUserMedia(function(err, stream) {
      if (this._state === STATE_REQUEST_CANCELLED) {
        this._state = STATE_NOT_RUNNING;
        return;
      }
      if (err !== null) {
        this._state = STATE_NOT_RUNNING;
        this.emit('error', err);
      } else {
        this._run(stream);
      }
    }.bind(this));
  };

  WordStream.prototype._run = function(stream) {
    this._state = STATE_RUNNING;
    this._stream = stream;

    var AudioContext = (window.AudioContext || window.webkitAudioContext);
    var context = new AudioContext();

    this._analyser = context.createAnalyser();
    this._analyser.fft = 2048;
    this._analyser.smoothingTimeConstant = 0;

    var source = context.createMediaStreamSource(stream);
    source.connect(this._analyser);

    this._interval = setInterval(this._tick.bind(this), SAMPLE_INTERVAL);
  };

  WordStream.prototype._stopRunning = function() {
    var tracks = this._stream.getAudioTracks();
    for (var i = 0, len = tracks.length; i < len; ++i) {
      tracks[i].stop();
    }
    this._stream = null;
    this._state = STATE_NOT_RUNNING;
  };

  WordStream.prototype._tick = function() {
    var bins = new Uint8Array(this._analyser.frequencyBinCount);
    this._analyser.getByteFrequencyData(bins);

    this._previousFFTs.push(bins);
    if (this._previousFFTs.length > FFT_BUFFER_SIZE) {
      this._previousFFTs.splice(0, 1);
    }

    var amplitude = 0;
    for (var i = 0, len = bins.length; i < len; ++i) {
      amplitude += Math.pow(bins[i], 2);
    }
    amplitude = Math.sqrt(amplitude);
    this._previousAmplitudes.push(amplitude);
    if (this._previousAmplitudes.length > AMPLITUDE_BUFFER_SIZE) {
      this._previousAmplitudes.splice(0, 1);
    } else {
      return;
    }

    var averageAmplitude = 0;
    for (var i = 0, len = this._previousAmplitudes.length; i < len; ++i) {
      averageAmplitude += this._previousAmplitudes[i];
    }
    averageAmplitude /= this._previousAmplitudes.length;

    var recentAverage = 0;
    for (var i = 0; i < AMPLITUDE_STOP_COUNT; ++i) {
      var idx = i + this._previousAmplitudes.length - AMPLITUDE_STOP_COUNT;
      recentAverage += this._previousAmplitudes[idx];
    }
    recentAverage /= AMPLITUDE_STOP_COUNT;

    if (recentAverage < averageAmplitude-400) {
      this._terminateWord();
    }
  };

  WordStream.prototype._terminateWord = function() {
    var signature = new window.app.WordSignature(this._previousFFTs);
    this._previousFFTs = [];
    this._previousAmplitudes = [];
    this.emit('word', signature);
  };

  window.app.WordStream = WordStream;

  function getUserMedia(cb) {
    var gum = (navigator.getUserMedia || navigator.webkitGetUserMedia ||
      navigator.mozGetUserMedia || navigator.msGetUserMedia);
    if (!gum) {
      setTimeout(function() {
        cb('getUserMedia() is not available.', null);
      }, 10);
      return;
    }
    gum.call(navigator, {audio: true, video: false},
      function(stream) {
        cb(null, stream);
      },
      function(err) {
        cb(err, null);
      }
    );
  }

})();
