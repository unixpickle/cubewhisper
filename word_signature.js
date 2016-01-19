(function() {

  function WordSignature(ffts) {
    this._vector = sumVectors(ffts);
    normalizeVector(this._vector);
  }

  WordSignature.prototype.difference = function(sig) {
    if (sig._vector.length !== this._vector.length) {
      throw new Error('signature vectors had different length');
    }
    var sum = 0;
    for (var i = 0, len = this._vector.length; i < len; ++i) {
      sum += Math.pow(sig._vector[i]-this._vector[i], 2);
    }
    return Math.sqrt(sum);
  };

  window.app.WordSignature = WordSignature;

  function sumVectors(vectors) {
    var sum = [];
    for (var i = 0, len = vectors.length; i < len; ++i) {
      var vector = vectors[i];
      for (var j = 0, len1 = vector.length; j < len1; ++j) {
        sum[j] = (sum[j] || 0) + vector[j];
      }
    }
    return sum;
  }

  function normalizeVector(vector) {
    var mag = vectorMagnitude(vector);
    if (mag === 0) {
      return;
    }
    for (var i = 0, len = vector.length; i < len; ++i) {
      vector[i] /= mag;
    }
  }

  function vectorMagnitude(vector) {
    var magSquared = 0;
    for (var i = 0, len = vector.length; i < len; ++i) {
      magSquared += Math.pow(vector[i], 2);
    }
    return Math.sqrt(magSquared);
  }

})();
