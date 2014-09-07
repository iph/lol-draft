library draft;
import 'dart:html';
import "netsocket.dart";
import 'package:angular/angular.dart';
import 'package:angular/application_factory_static.dart';
import 'angulardraft_generated_type_factory_maps.dart' show setStaticReflectorAsDefault;
import 'angulardraft_static_expressions.dart' as generated_static_expressions;
import 'angulardraft_static_metadata.dart' as generated_static_metadata;

int ENTER_KEY_CODE = 13;

class BidUpdate {
  String bid;
  String team;
  
  BidUpdate(this.bid, this.team);
}
class Player {
  String Ign;
  String Id;
  String Name;
  List<String> Roles;
  int Score;
  
  Player(this.Ign, this.Id, this.Name, this.Roles, this.Score);
}
@Controller(
  selector: '[draft-controller]',
  publishAs: 'ctrl')
class DraftController{
  NetworkSocket socket;
  DraftController(){
    Map map = splitKeys(window.location.search);
    token = map["Token"];
    NetworkSocket sock = new NetworkSocket(document.domain, "9001", (Map e){
      print("Val: " + e.toString());
      if(e["Type"] == "BID_UPDATE") {
        bids.add(new BidUpdate(e['Payload']["Bid"], e['Payload']['Team']));
        if(bids.length > 8) {
          bids.removeAt(0);
        }
      }
      if(e["Type"] == "BID_NEW") {
        bids.clear();
      }
      if(e["Type"] == "PLAYER_INFO") {
       
        Map p = e["Payload"]["Player"];
        print(p["Ign"]);
        player = new Player(p["Ign"], p["Id"], p["Name"], p["Roles"], p["Score"]);
        status = "...waiting to start bid";
      }
      
      if(e["Type"] == "WINNER"){
        status = "Sold: " + e["Payload"]["Team"];
        bids.clear();
      }
      
      if(e["Type"] == "COUNTDOWN"){
        if(e["Payload"] == "0"){
          status = "bidding";
        }else{
        var amt = int.parse(e["Payload"]);
        if(amt > 5) {
          status = "" + (10 - amt).toString();
        }
        }
        
      }
      
      if(e["Type"] == "BIDDER_INFO"){
        
      }
      
    });
    socket = sock;
  
  }
  
  keyPressed(KeyboardEvent event){
    if(event.keyCode == ENTER_KEY_CODE && !tooMuch){
      submitBid(int.parse(bidAmount));
    }
  }
  
  submitBid(amount) {
    bidAmount = "";
    socket.sendBid(amount, token);
  }
  
  String status;
  
  bool get tooMuch  => bidAmount.isEmpty|| int.parse(bidAmount, onError:(e) => null) == null || int.parse(bidAmount) > amount;
  int get remaining => amount - int.parse(bidAmount, onError:(e)=>0);
  int amount= 100;
  String token = "";
  String bidAmount = ""; 
  List<BidUpdate> bids = [];
  Player player;
}

class DraftModule extends Module {
  DraftModule(){
    bind(DraftController);
  }
}

Map<String, String> splitKeys(String search){
  Map<String, String> map = new Map<String, String>();
  if(search.length == 0){ return map;}
  var strs = search.substring(1).split("&");

  strs.forEach((str){
    var values = str.split("=");
    map[values[0]] = values[1];
  });
  
  return map;
}

void main() {
  setStaticReflectorAsDefault();
  Map map = splitKeys(window.location.search);
  if(!map.containsKey("Token")){
    print("Should fail here");
    //return;
  }
  staticApplicationFactory(generated_static_metadata.typeAnnotations, generated_static_expressions.getters, generated_static_expressions.setters, generated_static_expressions.symbols)
  ..addModule(new DraftModule())
  ..run();
}
