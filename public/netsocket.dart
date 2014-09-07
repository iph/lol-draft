
import "dart:html";
import "dart:convert";
import "dart:async";


class NetworkSocket {
  WebSocket websock;
  String port;
  String domain;
  var callback;
  NetworkSocket( this.domain, this.port, callback) {
    String uri = 'ws://${domain}:${port}/draft/';
    websock = new WebSocket(uri);
    websock.onOpen.listen((e){print("Open: " + e.toString());});
    
    websock.onError.listen((e){ print("Err:" + e.toString()); });
    
    var transformer = new StreamTransformer.fromHandlers(handleData: (MessageEvent value, sink){
      // Transform to a readable fucking format.
      FileReader reader = new FileReader();
      reader.onLoadEnd.listen((e){
        String jsonblob = UTF8.decode(reader.result);
        Map data = JSON.decode(jsonblob);
        sink.add(data);
      });
      reader.readAsArrayBuffer(value.data);
      
    });
    websock.onMessage.transform(transformer).listen((e) {
      callback(e);
    });
  }
  
  sendBid(int amount, String token){
    websock.sendString('{"Type":"BID", "Payload":"${amount}", "Token":"${token}"}');
  }
  
 
}