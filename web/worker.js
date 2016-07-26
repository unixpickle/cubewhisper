(function() {

  var RNN_FILE = 'speech_rnn';
  var WEBWORKER_FILE = 'webworker/webworker.js';

  var XHR_DONE = 4;
  var HTTP_OK = 200;

  function createWorker(callback) {
    fetchRNNData(function(err, rnnData) {
      if (err) {
        callback(err, null);
      } else {
        callback(null, new WorkerWrapper(rnnData));
      }
    });
  }

  window.app.createWorker = createWorker;

  function fetchRNNData(callback) {
    var xhr = new XMLHttpRequest();
    xhr.responseType = "arraybuffer";
    xhr.open('GET', RNN_FILE);
    xhr.send(null);

    xhr.onreadystatechange = function () {
      if (xhr.readyState === XHR_DONE) {
        if (xhr.status === HTTP_OK) {
          callback(null, new Uint8Array(xhr.response));
        } else {
          callback('Error: '+xhr.status, null);
        }
      }
    };
  }

  function WorkerWrapper(rnnData) {
    window.EventEmitter.call(this);
    this._currentWorker = new Worker(WEBWORKER_FILE);
    this._currentWorker.onmessage = function(e) {
      if ('undefined' !== typeof e.data.status) {
        this.emit('loading', e.data.status);
      } else {
        this.emit('moves', e.data.moves, e.data.raw);
      }
    }.bind(this);
    this._currentWorker.postMessage(['init', rnnData]);
  }

  WorkerWrapper.prototype = Object.create(window.EventEmitter.prototype);
  WorkerWrapper.prototype.constructor = WorkerWrapper;

  WorkerWrapper.prototype.start = function(rate) {
    this._currentWorker.postMessage(['start', rate]);
  };

  WorkerWrapper.prototype.samples = function(s) {
    this._currentWorker.postMessage(['samples', s]);
  };

  WorkerWrapper.prototype.end = function() {
    this._currentWorker.postMessage(['end']);
  };

})();
