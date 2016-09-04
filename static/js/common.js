define(function() {

var getUrlVars = function (){
    var vars = [], hash;
    var hashes = window.location.href.slice(window.location.href.indexOf('?') + 1).split('&');
    for(var i = 0; i < hashes.length; i++)
    {
      hash = hashes[i].split('=');
      vars.push(hash[0]);
      vars[hash[0]] = hash[1];
    }
    return vars;
}

var getUrlVar = function (name){
    return getUrlVars()[name];
}

var arrayBufferToBase64 = function ( buffer ) {
  var binary = '';
  var bytes = new Uint8Array( buffer );
  var len = bytes.byteLength;
  for (var i = 0; i < len; i++) {
      binary += String.fromCharCode( bytes[ i ] );
  }
  return window.btoa( binary );
}

var base64ToArrayBuffer = function (base64) {
    var binary_string =  window.atob(base64);
    var len = binary_string.length;
    var bytes = new Uint8Array( len );
    for (var i = 0; i < len; i++)        {
        bytes[i] = binary_string.charCodeAt(i);
    }
    return bytes.buffer;
}

return {
  getUrlVars : getUrlVars,
  getUrlVar : getUrlVar,
  arrayBufferToBase64 : arrayBufferToBase64,
  base64ToArrayBuffer : base64ToArrayBuffer
};

});

