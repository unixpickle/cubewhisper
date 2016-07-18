(function() {

  var STATE_NOT_RUNNING = 0;
  var STATE_REQUESTED = 1;
  var STATE_REQUEST_CANCELLED = 2;
  var STATE_RUNNING = 3;

  function SampleStream() {
    window.EventEmitter.call(this);
    this._state = STATE_NOT_RUNNING;
    this._stream = null;
    this._context = null;
    this._source = null;
    this._node = null;
  }

  SampleStream.prototype = Object.create(window.EventEmitter.prototype);
  SampleStream.prototype.constructor = SampleStream;

  SampleStream.prototype.start = function() {
    switch (this._state) {
    case STATE_NOT_RUNNING:
      this._requestUserMedia();
      break;
    case STATE_REQUEST_CANCELLED:
      this._state = STATE_REQUESTED;
      break;
    }
  };

  SampleStream.prototype.stop = function() {
    switch (this._state) {
    case STATE_RUNNING:
      this._stopRunning();
      break;
    case STATE_REQUESTED:
      this._state = STATE_REQUEST_CANCELLED;
      break;
    }
  };

  SampleStream.prototype._requestUserMedia = function() {
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

  SampleStream.prototype._run = function(stream) {
    this._state = STATE_RUNNING;
    this._stream = stream;

    this._context = getAudioContext();
    this._source = this._context.createMediaStreamSource(stream);
    this._node = new SampleNode(this._context, this);
    this._source.connect(this._node.node());
    this._node.node().connect(this._context.destination);
  };

  SampleStream.prototype._stopRunning = function() {
    var tracks = this._stream.getAudioTracks();
    for (var i = 0, len = tracks.length; i < len; ++i) {
      tracks[i].stop();
    }
    this._source.disconnect(this._node.node());
    this._node.node().disconnect(this._context.destination);

    this._stream = null;
    this._context = null;
    this._source = null;
    this._node = null;
    this._state = STATE_NOT_RUNNING;
  };

  window.app.SampleStream = SampleStream;

  function SampleNode(context, emitter) {
    this._emitter = emitter;
    this._started = false;
    this._sampleCount = 0;
    this._sampleRate = 0;
    if (context.createScriptProcessor) {
      this._node = context.createScriptProcessor(1024, 1, 1);
    } else if (context.createJavaScriptNode) {
      this._node = context.createJavaScriptNode(1024, 1, 1);
    } else {
      throw new Error('No javascript processing node available.');
    }
    this._node.onaudioprocess = this._process.bind(this);
  }

  SampleNode.prototype.node = function() {
    return this._node;
  };

  SampleNode.prototype._process = function(event) {
    var input = event.inputBuffer;
    if (!this._started) {
      this._started = true;
      this._emitter.emit('start', Math.round(input.sampleRate));
    }

    var sampleCount = input.length;
    var channels = input.numberOfChannels;

    var monoSamples = [];
    for (var channel = 0; channel < channels; ++channel) {
      var samples = input.getChannelData(channel);
      for (var i = 0; i < sampleCount; ++i) {
        if (channel === 0) {
          monoSamples[i] = samples[i];
        } else {
          monoSamples[i] += samples[i];
        }
      }
    }

    for (var i = 0; i < sampleCount; ++i) {
      monoSamples[i] /= channels;
    }

    this._emitter.emit('samples', monoSamples);
  };

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

  var reusableAudioContext = null;

  function getAudioContext() {
    if (reusableAudioContext !== null) {
      return reusableAudioContext;
    }
    var AudioContext = (window.AudioContext || window.webkitAudioContext);
    reusableAudioContext = new AudioContext();
    return reusableAudioContext;
  }

})();
