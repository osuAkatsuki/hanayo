$(document).ready(function () {
  var wl = window.location;
  var newPathName = wl.pathname + wl.search;

  if (newPathName.split("/")[2] != clanID) {
    newPathName = "/c/" + clanID + wl.search;
  }

  // todo: same for relax check /// build proper path (using doubled replaceState can confuse users)
  var b = wl.search.length !== 0;
  if (wl.search.indexOf("mode=") === -1) {
    newPathName += (b ? "&" : "?") + "mode=" + favouriteMode;
    b = true;
  }
  if (wl.search.indexOf("rx=") === -1) {
    newPathName += (b ? "&" : "?") + "rx=" + rx;
    b = true;
  }

  /*if (!b && wl.pathname != newPathName)
		window.history.replaceState('', document.title, newPathName + wl.search + wl.hash);
	else*/
  window.history.replaceState("", document.title, newPathName + wl.hash);

  /*if (wl.search.indexOf("rx=") === -1) {

	}*/
  toggleModeAvailability(favouriteMode, rx);
  setMode(favouriteMode, rx);

  $("#rx-menu>.item").on("click", function (e) {
    e.preventDefault();
    if ($(this).hasClass("active")) return;

    var nrx = $(this).data("rx");
    $("#rx-menu>.active.item").removeClass("active");
    window.rx = nrx;
    toggleModeAvailability(favouriteMode, rx);
    $("[data-mode]:not(.item):not([hidden])").attr("hidden", "");
    $(
      "[data-mode=" +
        favouriteMode +
        (rx != 0 ? (rx != 2 ? "r" : "a") : "") +
        "]:not(.item)"
    ).removeAttr("hidden");
    setMode(favouriteMode, rx);
    $(this).addClass("active");
    window.history.replaceState(
      "",
      document.title,
      wl.pathname + "?mode=" + favouriteMode + "&rx=" + nrx + wl.hash
    );
  });

  $("#mode-menu>.item").on("click", function (e) {
    e.preventDefault();
    if ($(this).hasClass("active")) return;

    var m = $(this).data("mode");
    $("#mode-menu>.active.item").removeClass("active");
    //todo: let new stats table show and hide old
    toggleModeAvailability(m, rx);
    window.favouriteMode = m;
    $("[data-mode]:not(.item):not([hidden])").attr("hidden", "");
    $(
      "[data-mode=" +
        m +
        (rx != 0 ? (rx != 2 ? "r" : "a") : "") +
        "]:not(.item)"
    ).removeAttr("hidden");
    setMode(m, rx);
    $(this).addClass("active");
    window.history.replaceState(
      "",
      document.title,
      wl.pathname + "?mode=" + m + "&rx=" + rx + wl.hash
    );
  });

  $("#join-btn>.item").on("click", function (e) {
    e.preventDefault();
    if (!currentUserID) return;

    var btn = $(this);
    joinClan({ id: parseInt(clanID) }, btn);
  });

  $("#leave-btn>.item").on("click", function (e) {
    e.preventDefault();
    if (!currentUserID) return;
    var thiss = $(this);
    api(
      "clans/leave",
      null,
      function (t) {
        if (t.message === "disbanded") {
          location.replace("/");
        } else {
          location.replace("/c/" + clanID);
        }
        showMessage("success", "Successfully left.");
      },
      !0
    );
  });
});

function joinClan(obj, btn) {
  api(
    "clans/join",
    obj,
    function (t) {
      if (t.code == 403) {
        showMessage("error", "You are not authorized to do this.");
        return;
      } else if (t.code == 400) {
        showMessage("error", "Invalid Invite Format.");
        return;
      } else if (t.code == 404) {
        showMessage("error", "Invite was not found.");
        return;
      }

      btn.text("Leave");
      btn.attr("id", "leave-btn");
      btn.css("background-color", "rgba(255,0,0,.3)");
      btn.unbind();
      showMessage("success", "Successfully joined " + t.clan.name);
      api("users", { id: "self" }, function (r) {
        document.getElementById(
          "members"
        ).innerHTML += `<div class="column"> <div class="ui left aligned fluid card"> <div class="image"> <img src="${
          hanayoConf.avatars
        }/${
          r.id
        }" alt="Avatar"> </div> <div class="content"> <a class="header" href="/u/"><i class="${r.country.toLowerCase()} flag"></i>${
          r.username
        }</a> </div> </div> </div>`;
      });
      setMode(favouriteMode, rx);
    },
    !0
  );
}

function setMode(mode, rx) {
  let eldx = document.getElementById(
    mode + (rx != 0 ? (rx != 2 ? "r" : "a") : "")
  );
  api("clans/stats", { id: clanID, m: mode, rx: rx }, function (e) {
    var data = e.clan.chosen_mode;
    eldx.innerHTML =
      `` +
      tableRow("Global Rank", addCommas(data.global_leaderboard_rank)) +
      tableRow("Performance", addCommas(data.pp) + "pp") +
      tableRow("Ranked Score", addCommas(data.ranked_score)) +
      tableRow("Total Score", addCommas(data.total_score)) +
      tableRow("Total Playcount", addCommas(data.playcount)) +
      tableRow("Total Replays Watched", addCommas(data.replays_watched)) +
      tableRow("Total Hits", addCommas(data.total_hits));
  });
}

function tableRow(title, data) {
  return `<tr><td><b>${title}</b></td> <td class="right aligned">${data}</td></tr>`;
}

function toggleModeAvailability(mode, rx) {
  for (i = 0; i <= 3; i++) {
    $(`[data-mode='${i}']`).removeClass("disabled");
  }

  for (i = 0; i <= 2; i++) {
    $(`[data-rx='${i}']`).removeClass("disabled");
  }

  if (rx == 1) {
    // relax does not have mania
    $("[data-mode='3']").addClass("disabled");
  } else if (rx == 2) {
    // autopilot does not have taiko, catch, or mania
    $("[data-mode='1']").addClass("disabled");
    $("[data-mode='2']").addClass("disabled");
    $("[data-mode='3']").addClass("disabled");
  }

  if (mode == 1 || mode == 2) {
    // taiko or catch does not have autopilot
    $("[data-rx='2']").addClass("disabled");
  } else if (mode == 3) {
    // mania does not have relax or autopilot
    $("[data-rx='1']").addClass("disabled");
    $("[data-rx='2']").addClass("disabled");
  }
}
