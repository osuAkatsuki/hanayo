/*
//////////////////////////////////////////
//////////Featured Video Grabber//////////
////////Grab the Latest Video from////////
//////////Atoka Replays Channel///////////
//////////////////////////////////////////
*/

var channelID = "UCNuP9DpFmwem0GLN-0QsIlA";
var reqURL = "https://www.youtube.com/feeds/videos.xml?channel_id=";
$.getJSON("https://api.rss2json.com/v1/api.json?rss_url=" + encodeURIComponent(reqURL)+channelID, function(data) {
   var link = data.items[0].link;
   var id = link.substr(link.indexOf("=")+1);
$("#featured_video").attr("src","https://youtube.com/embed/"+id + "?controls=1&showinfo=1&rel=0");
});

// This checks for the latest video from the Atoka Replays Channel.
// To change the channel that it checks for just copy the channel you want's id and paste it in the channelID variable.
// Enable or Disable Control by changing controls variable in line 12. 1 = on. 0 = off.


// The second part of the code that usually goes in a HTML document looks like this:
/*
<script src="https://ajax.googleapis.com/ajax/libs/jquery/2.1.1/jquery.min.js"></script>
<script src="../static/js/featured_video.js"></script>
<iframe id="featured_video" width="600" height="340" frameborder="0" allowfullscreen></iframe>
*/
