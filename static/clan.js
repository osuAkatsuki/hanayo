var modesSet = new Array(7).fill(false);
$(document).ready(function() {
	var wl = window.location;
	var newPathName = wl.pathname;

	if (newPathName.split('/')[2] != clanID) {
		newPathName = "/c/" + clanID;
	}
	
	// todo: same for relax check /// build proper path (using doubled replaceState can confuse users)
	var b = false;
	if (wl.search.indexOf("mode=") === -1) {
		newPathName += "?mode=" + favouriteMode;
		b = true;
	}
	if (wl.search.indexOf("rx=") === -1) {
		newPathName += (b ? '&' : '?') + "rx=" + rx;
		b = true;
	}
		
	if (!b && wl.pathname != newPathName)
		window.history.replaceState('', document.title, newPathName + wl.search + wl.hash);
	else 
		window.history.replaceState('', document.title, newPathName + wl.search + wl.hash);
	
	/*if (wl.search.indexOf("rx=") === -1) {
		
	}*/
	setMode(favouriteMode, rx != 0);
	$("#rx-menu>.item").click(function(e) {
		e.preventDefault();
		if ($(this).hasClass("active")) return;
		var nrx = $(this).data("rx");
		$("#rx-menu>.active.item").removeClass("active");
		window.rx = nrx;
		$("[data-mode]:not(.item):not([hidden])").attr("hidden", "");
		$("[data-mode=" + favouriteMode + (rx != 0 ? 'r' : '') + "]:not(.item)").removeAttr("hidden");
		setMode(favouriteMode, rx != 0);
		$(this).addClass("active");
		window.history.replaceState('', document.title, wl.pathname + "?mode=" + favouriteMode + "&rx=" + nrx + wl.hash);
	});
		
	$("#mode-menu>.item").click(function(e) {
		e.preventDefault();
		if ($(this).hasClass("active")) return;
		
		var m = $(this).data("mode");
		$("#mode-menu>.active.item").removeClass("active");
		//todo: let new stats table show and hide old
		window.favouriteMode = m;
		$("[data-mode]:not(.item):not([hidden])").attr("hidden", "");
		$("[data-mode=" + m + (rx != 0 ? 'r' : '') + "]:not(.item)").removeAttr("hidden");
		setMode(m, rx != 0);
		$(this).addClass("active");
		window.history.replaceState('', document.title, wl.pathname + "?mode=" + m + "&rx=" + rx + wl.hash);
	})
});

function setMode(mode, rx) {
	var mIndex = rx ? mode + 4 : mode;
	if (mIndex > 6 || mIndex < 0) return;
	if (modesSet[mIndex]) return;
	modesSet[mIndex] = true;
	eldx = document.getElementById(mode + (rx ? 'r' : ''));
	eldx.innerHTML = "Mode: " + mode;
	
	api("clans/stats", { id: clanID, m: mode, rx: (rx ? 1 : 0) }, function (e) {
		var data = e.clan.chosen_mode;
		eldx.innerHTML = `<td></td>` + tableRow("Global Rank", addCommas(data.global_leaderboard_rank)) 
		+ tableRow("Performance", addCommas(data.pp)+"pp") 
		+ tableRow("Ranked Score", data.ranked_score)
		+ tableRow("Total Score", data.total_score)
		+ tableRow("Total Playcount", data.playcount)
		+ tableRow("Total Replays Watched", data.replays_watched)
		+ tableRow("Total Hits", data.total_hits);
	});
}

function tableRow(title, data) {
	return `<tr><td><b>${title}</b></td> <td class="right aligned">${data}</td></tr>`;
}
