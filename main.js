window.addEventListener('load', function() {
  var stream = new window.app.WordStream();
  stream.on('error', function(err) {
    alert('Error: ' + err);
  });
  stream.start();

  var el = document.createElement('div');
  el.innerText = '0 words';
  var wordCount = 0;
  document.body.appendChild(el);
  var lastSignature = null;
  stream.on('word', function(sig) {
    ++wordCount;
    el.innerText = wordCount + ' word' + (wordCount === 1 ? '' : 's');
    if (lastSignature !== null) {
      el.innerText += ' difference from last: ' + sig.difference(lastSignature);
    }
    lastSignature = sig;
  });
});
