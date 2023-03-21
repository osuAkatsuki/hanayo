(function () {
  var mapset = {};
  setData.ChildrenBeatmaps.forEach(function (diff) {
    mapset[diff.BeatmapID] = diff;
  });
  console.log(mapset);
  function loadLeaderboard(b, m, rx) {
    var wl = window.location;
    window.history.replaceState(
      "",
      document.title,
      "/b/" + b + "?mode=" + m + "&rx=" + rx + wl.hash
    );
    var Score = rx ? "pp" : "score";
    api(
      `scores?sort=${Score},desc`,
      {
        m: m,
        relax: rx,
        b: b,
        p: 1,
        l: 50,
      },
      function (data) {
        console.log(data);
        var tb = $(".ui.table tbody");
        tb.find("tr").remove();
        if (data.scores == null) {
          data.scores = [];
        }
        var i = 0;
        data.scores.sort(function (a, b) {
          return b.Score - a.Score;
        });
        data.scores.forEach(function (score) {
          var user = score.user;
          var scoreRank = getRank(
            m,
            score.mods,
            score.accuracy,
            score.count_300,
            score.count_100,
            score.count_50,
            score.count_miss
          );
          var scoreRankIcon = `<a class="rank-${scoreRank} rank--bmap--font">${scoreRank.replace(
            "HD",
            "+"
          )}</a>`;
          tb.append(
            $("<tr class='l-player' />").append(
              $("<td data-sort-value=" + ++i + " />").html(
                `#${(page - 1) * 50 + i}`
              ),
              $("<td />").html(
                "<a href='/u/" +
                  user.id +
                  "' title='View profile'><img src='/static/images/flags/" +
                  countryToCodepoints(user.country.toUpperCase()) +
                  ".svg' class='new-flag fixed--flag--margin'></img>" +
                  " " +
                  escapeHTML(user.username) +
                  "</a>"
              ),
              $("<td class='center aligned' />").html(scoreRankIcon),
              $("<td class='center aligned' data-sort-value=" + score.score + " />").html(
                addCommas(score.score)
              ),
              $("<td class='center aligned' data-sort-value=" + score.accuracy + " />").text(
                score.accuracy.toFixed(2) + "%"
              ),
              $("<td class='center aligned' data-sort-value=" + score.max_combo + " />").text(
                addCommas(score.max_combo) + "x"
              ),
              $("<td class='center aligned' data-sort-value=" + score.pp.toFixed(0) + " />").html(
                score.pp.toFixed(0) + "pp"
              ),
              $("<td class='center aligned' />").html(getScoreMods(score.mods, true)),
              $("<td class='center aligned' data-sort-value=" + Date.parse(score.time).valueOf() + " />").html(timeSince(Date.parse(score.time))),
              $("<td class='center aligned' />").html(
                "<a href='/web/replays/" + 
                score.id + 
                "' title='Download Replay' class='new downloadstar'><i class='fa-solid fa-download icon'></i>Get</a>"
              )
            )
          );
        });
      }
    );
  }
  function changeDifficulty(bid) {
    // load info
    var diff = mapset[bid];

    // column 2
    $("#cs").html(diff.CS);
    $("#hp").html(diff.HP);
    $("#od").html(diff.OD);
    $("#passcount").html(addCommas(diff.Passcount));
    $("#playcount").html(addCommas(diff.Playcount));

    // column 3
    $("#ar").html(diff.AR);
    $("#stars").html(diff.DifficultyRating.toFixed(2));
    $("#length").html(timeFormat(diff.TotalLength));
    $("#drainLength").html(timeFormat(diff.HitLength));
    $("#bpm").html(diff.BPM);

    // hide mode for non-std maps
    if (diff.Mode != 0) {
      currentMode = diff.Mode;
      $("#rx-column").removeClass("five wide column").addClass("sixteen wide column");
      $("#mode-column").hide();
    } else {
      if (currentMode === null) {
        currentMode = favMode;
      }
      $("#rx-column").removeClass("sixteen wide column").addClass("five wide column");
      $("#mode-column").show();
    }

    if (diff.Mode == 3) {
      currentCmode = 0;
      $("#mode-column").removeClass("eleven wide column").addClass("sixteen wide column");
      $("#rx-column").hide();
    } else {
      $("#mode-column").removeClass("sixteen wide column").addClass("eleven wide column");
      $("#rx-column").show();
    }

    // update mode menu
    $("#mode-menu .active.item").removeClass("active");
    $("#mode-" + currentMode).addClass("active");

    // update cmode menu
    $("#cmode-menu .active.item").removeClass("active");
    $("#cmode-" + currentCmode).addClass("active");

    loadLeaderboard(bid, currentMode, currentCmode);
    toggleModeAvailability(currentMode, currentCmode);
  }
  window.loadLeaderboard = loadLeaderboard;
  window.changeDifficulty = changeDifficulty;
  changeDifficulty(beatmapID);
  // loadLeaderboard(beatmapID, currentMode);
  $("#diff-menu .item").on("click", function (e) {
    e.preventDefault();
    $(this).addClass("active");
    beatmapID = $(this).data("bid");
    changeDifficulty(beatmapID);
  });
  $("#mode-menu .item").on("click", function (e) {
    e.preventDefault();
    $("#mode-menu .active.item").removeClass("active");
    $(this).addClass("active");
    currentMode = $(this).data("mode");
    loadLeaderboard(beatmapID, currentMode, currentCmode);
    toggleModeAvailability(currentMode, currentCmode);
    currentModeChanged = true;
  });
  $("#cmode-menu .item").on("click", function (e) {
    e.preventDefault();
    $("#cmode-menu .active.item").removeClass("active");
    $(this).addClass("active");
    currentCmode = $(this).data("cmode");
    loadLeaderboard(beatmapID, currentMode, currentCmode);
    toggleModeAvailability(currentMode, currentCmode);
    currentCmodeChanged = true;
  });
  $("table.sortable").tablesort();
})();

function getRank(gameMode, mods, acc, c300, c100, c50, cmiss) {
  var total = c300 + c100 + c50 + cmiss;

  // Hidden | Flashlight | FadeIn
  var hdfl = (mods & 1049608) > 0;

  var ss = hdfl ? "SSHD" : "SS";
  var s = hdfl ? "SHD" : "S";

  switch (gameMode) {
    case 0:
    case 1:
      var ratio300 = c300 / total;
      var ratio50 = c50 / total;

      if (ratio300 == 1) return ss;

      if (ratio300 > 0.9 && ratio50 <= 0.01 && cmiss == 0) return s;

      if ((ratio300 > 0.8 && cmiss == 0) || ratio300 > 0.9) return "A";

      if ((ratio300 > 0.7 && cmiss == 0) || ratio300 > 0.8) return "B";

      if (ratio300 > 0.6) return "C";

      return "D";

    case 2:
      if (acc == 100) return ss;

      if (acc > 98) return s;

      if (acc > 94) return "A";

      if (acc > 90) return "B";

      if (acc > 85) return "C";

      return "D";

    case 3:
      if (acc == 100) return ss;

      if (acc > 95) return s;

      if (acc > 90) return "A";

      if (acc > 80) return "B";

      if (acc > 70) return "C";

      return "D";
  }
}
function toggleModeAvailability(mode, rx) {
  $("[data-mode='1']").removeClass("disabled");
  $("[data-mode='2']").removeClass("disabled");
  $("[data-mode='3']").removeClass("disabled");

  $("[data-cmode='1']").removeClass("disabled");
  $("[data-cmode='2']").removeClass("disabled");
  $("[data-cmode='3']").removeClass("disabled");

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
    $("[data-cmode='2']").addClass("disabled");
  } else if (mode == 3) {
    // mania does not have relax or autopilot
    $("[data-cmode='1']").addClass("disabled");
    $("[data-cmode='2']").addClass("disabled");
  }
}
