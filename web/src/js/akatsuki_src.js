/*!
 * akatsuki_src.js
 * Copyright (C) 2016 Morgan Bazalgette and Giuseppe Guerra
 * Copyright (C) 2017 Akatsuki
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

// this object contains tiny snippets that were deemed too small to be worth
// their own file.

const PREMIUM_PRICE_PER_MONTH = 5.0;

var singlePageSnippets = {
  "/clans": function () {
    page = 0 === page ? 1 : page;

    function buildClanPPItem(v, idx) {
      return $("<tr class='l-clan' />").append(
        $("<td />").text("#" + (50 * (page - 1) + idx)),
        $("<td />").html(
          "<a href='/c/" +
            v.id +
            "?mode=" +
            favouriteMode +
            "&rx=" +
            rx +
            "' title='View clan'>" +
            // TODO: uncomment it once we have the tag in the API response
            // escapeHTML(`[${v.tag}] ${v.name}`) +
            escapeHTML(v.name) +
            "</a>"
        ),
        $("<td class='center aligned' />").html(
          scoreOrPP(v.chosen_mode.ranked_score, v.chosen_mode.pp)
        ),
        $("<td class='center aligned' />").text(
          v.chosen_mode.accuracy.toFixed(2) + "%"
        ),
        $("<td class='center aligned' />").html(
          addCommas(v.chosen_mode.playcount)
        )
      );
    }

    function buildClan1sItem(v, idx) {
      return $("<tr class='l-clan' />").append(
        $("<td />").text("#" + (50 * (page - 1) + idx)),
        $("<td />").html(
          "<a href='/c/" +
            v.clan +
            "?mode=" +
            favouriteMode +
            "&rx=" +
            rx +
            "' title='View clan'>" +
            escapeHTML(`[${v.tag}] ${v.name}`) +
            "</a>"
        ),
        $("<td class='center aligned' />").html(addCommas(v.count))
      );
    }

    function loadClanLeaderboard() {
      var wl = window.location;
      window.history.replaceState(
        "",
        document.title,
        wl.pathname +
          "?mode=" +
          favouriteMode +
          "&p=" +
          page +
          "&rx=" +
          rx +
          "&sort=" +
          sort +
          wl.hash
      );

      apiRoute = sort == "1s" ? "first" : "all";

      api(
        `clans/stats/${apiRoute}`,
        {
          m: favouriteMode,
          p: page,
          l: 50,
          rx: rx,
        },
        function (data) {
          var tb = $(".ui.table tbody");
          tb.find("tr").remove();
          if (data.clans == null) {
            disableSimplepagButtons(true);
            data.clans = [];
          }

          var i = 0;
          data.clans.forEach(function (v) {
            i++; // cursed
            if (sort == "1s") {
              clanItem = buildClan1sItem(v, i);
            } else {
              clanItem = buildClanPPItem(v, i);
            }

            tb.append(clanItem);
          });
          disableSimplepagButtons(data.clans.length < 50);
        }
      );
    }

    function scoreOrPP(s, pp) {
      if (pp === 0) {
        return "<b>" + addCommas(s) + "</b>";
      }

      return "<b>" + addCommas(pp) + "pp</b> (" + addCommas(s) + ")";
    }

    $("#rx-menu ." + rx + ".item").addClass("active");
    $("#sort-menu ." + ("1s" === sort ? 0 : 1) + ".item").addClass("active");
    toggleModeAvailability(favouriteMode, rx);

    loadClanLeaderboard();
    setupSimplepag(loadClanLeaderboard);

    $("#mode-menu .item").on("click", function (e) {
      e.preventDefault();
      $("#mode-menu .active.item").removeClass("active");
      $(this).addClass("active");
      favouriteMode = $(this).data("mode");
      page = 1;
      toggleModeAvailability(favouriteMode, rx);
      loadClanLeaderboard();
    });

    $("#rx-menu .item").on("click", function (e) {
      e.preventDefault();
      $("#rx-menu .active.item").removeClass("active");
      $(this).addClass("active");
      rx = $(this).data("rx");
      page = 1;
      toggleModeAvailability(favouriteMode, rx);
      loadClanLeaderboard();
    });
  },

  "/leaderboard": function () {
    page = page === 0 ? 1 : page;

    function loadLeaderboard() {
      var wl = window.location;
      window.history.replaceState(
        "",
        document.title,
        wl.pathname +
          "?mode=" +
          favouriteMode +
          "&rx=" +
          rx +
          "&p=" +
          page +
          "&sort=" +
          sort +
          (country != "" ? "&country=" + encodeURI(country) : "") +
          wl.hash
      );
      api(
        "leaderboard",
        {
          mode: favouriteMode,
          rx: rx,
          p: page,
          l: 50,
          country: country,
          sort: sort,
        },
        function (data) {
          var tb = $(".ui.table tbody");
          tb.find("tr").remove();
          if (data.users == null) {
            disableSimplepagButtons(true);
            data.users = [];
          }
          var i = 0;
          data.users.forEach(function (v) {
            codepoints = countryToCodepoints(v.country);
            tb.append(
              $("<tr class='l-player' />").append(
                $("<td />").text("#" + ((page - 1) * 50 + ++i)),
                $("<td />").html(
                  "<div class='flag-container'><a href='/leaderboard?mode=" +
                    favouriteMode +
                    "&rx=" +
                    rx +
                    "&country=" +
                    v.country.toLowerCase() +
                    "'><img src='/static/images/flags/" +
                    codepoints +
                    ".svg' class='new-flag nopad'></img></a><a href='/u/" +
                    v.id +
                    "?mode=" +
                    favouriteMode +
                    "&rx=" +
                    rx +
                    "'>" +
                    escapeHTML(v.username) +
                    "</a></div>"
                ),
                $("<td class='center aligned' />").html(
                  scoreOrPP(
                    v.chosen_mode.ranked_score,
                    v.chosen_mode.pp,
                    sort === "score"
                  )
                ),
                $("<td class='center aligned' />").text(
                  v.chosen_mode.accuracy.toFixed(2) + "%"
                ),
                $("<td class='center aligned' />").html(
                  addCommas(v.chosen_mode.playcount) +
                    " <i>(lv. " +
                    v.chosen_mode.level.toFixed(0) +
                    ")</i>"
                )
              )
            );
          });
          disableSimplepagButtons(data.users.length < 50);
        }
      );
    }

    function scoreOrPP(s, pp, forceScore) {
      if (pp === 0) {
        return "<b>" + addCommas(s) + "</b>";
      }

      if (forceScore) {
        return "<b>" + addCommas(s) + "</b> (" + addCommas(pp) + "pp)";
      }

      return "<b>" + addCommas(pp) + "pp</b> (" + addCommas(s) + ")";
    }

    if (country) {
      $(`[data-country=${country}]`).addClass("active");
    }

    // country stuff
    $("#country-chooser-modal").on("click", function () {
      $(".ui.modal").modal("show");
    });

    $(".lb-country").on("click", function () {
      country = $(this).data("country");
      $("#country-menu .active.item").removeClass("active");
      $(this).addClass("active");
      page = 1;
      $(".ui.modal").modal("hide");
      loadLeaderboard();
    });

    toggleModeAvailability(favouriteMode, rx);
    loadLeaderboard();
    setupSimplepag(loadLeaderboard);

    $("#mode-menu .item").on("click", function (e) {
      e.preventDefault();
      $("#mode-menu .active.item").removeClass("active");
      $("#country-menu .active.item").removeClass("active");
      $(this).addClass("active");
      favouriteMode = $(this).data("mode");
      country = "";
      page = 1;
      toggleModeAvailability(favouriteMode, rx);
      loadLeaderboard();
    });

    $("#rx-menu .item").on("click", function (e) {
      e.preventDefault();
      $("#rx-menu .active.item").removeClass("active");
      $("#country-menu .active.item").removeClass("active");
      $(this).addClass("active");
      country = "";
      page = 1;
      rx = $(this).data("rx");
      toggleModeAvailability(favouriteMode, rx);
      loadLeaderboard();
    });
  },

  "/friends": function () {
    $(".ui.compact.labeled.button").on("click", function () {
      var t = $(this);
      var delAdd = t.data("deleted") == 1 ? "add" : "del";
      t.addClass("disabled");
      api(
        "friends/" + delAdd,
        { user: parseInt(t.data("userid")) },
        function (data) {
          t.removeClass("disabled");
          t.data("deleted", data.friend ? "0" : "1");
          t.removeClass("green red blue");
          t.addClass(data.friend ? (data.mutual ? "red" : "green") : "blue");
          t.find(".icon")
            .removeClass("minus plus heart")
            .removeClass("green red blue")
            .addClass(
              data.friend
                ? data.mutual
                  ? "red heart"
                  : "green minus"
                : "blue plus"
            );
          t.find("span").text(data.friend ? T("Unfriend") : T("Befriend"));
        },
        true
      );
    });
  },

  "/followers": function () {
    $(".ui.compact.labeled.button").on("click", function () {
      var t = $(this);
      var delAdd = t.data("deleted") == 1 ? "add" : "del";
      t.addClass("disabled");
      api(
        "friends/" + delAdd,
        { user: parseInt(t.data("userid")) },
        function (data) {
          t.removeClass("disabled");
          t.data("deleted", data.friend ? "0" : "1");
          t.removeClass("green red");
          t.addClass(data.friend && data.mutual ? "red" : "green");
          t.find(".icon")
            .removeClass("plus heart")
            .removeClass("green red")
            .addClass(
              data.friend && data.mutual
                  ? "red heart"
                  : "green plus"
            );
          t.find("span").text(data.friend && data.mutual ? T("Unmutual") : T("Mutual"));
        },
        true
      );
    });
  },

  "/team": function () {
    $("#everyone").on("click", function () {
      $(".ui.modal").modal("show");
    });
  },

  "/register/verify": function () {
    var qu = query("u");
    setInterval(function () {
      $.getJSON(
        hanayoConf.banchoAPI + "/api/v1/verifiedStatus?u=" + qu,
        function (data) {
          if (data.result >= 0) {
            window.location.href = "/register/welcome?u=" + qu;
          }
        }
      );
    }, 5000);
  },

  "/settings": function () {
    $("input[name='custom_badge.icon']").on("input", function () {
      $("#badge-icon").attr(
        "class",
        "circular big icon " + escapeHTML($(this).val())
      );
    });
    $("input[name='custom_badge.name']").on("input", function () {
      $("#badge-name").html(escapeHTML($(this).val()));
    });
    $("input[name='custom_badge.show']").change(function () {
      if ($(this).is(":checked")) $("#custom-badge-fields").slideDown();
      else $("#custom-badge-fields").slideUp();
    });
    var isDark = $("#dark-site").is(":checked");
    $("form").submit(function (e) {
      e.preventDefault();

      var darkSetting = $("#dark-site").is(":checked");
      if (darkSetting != isDark) {
        var cflags = document.cookie.replace(
          /(?:(?:^|.*;\s*)cflags\s*\=\s*([^;]*).*$)|^.*$/,
          "$1"
        );
        cflags = darkSetting ? +cflags | 1 : +cflags & ~1;
        document.cookie = "cflags=" + cflags + ";path=/;max-age=31536000";
      }

      var obj = formToObject($(this));
      var ps = 0;
      $(this)
        .find("input[data-sv]")
        .each(function (_, el) {
          el = $(el);
          if (el.is(":checked")) {
            ps |= el.data("sv");
          }
        });
      obj.leaderboard_size = obj.leaderboard_size !== "" ? parseInt(obj.leaderboard_size) : hanayoConf.defaultLeaderboardSize
      obj.play_style = ps;
      var f = $(this);
      api(
        "users/self/settings",
        obj,
        function (data) {
          if (darkSetting != isDark) {
            window.location.reload();
            return;
          }
          showMessage("success", "Your new settings have been saved.");
          f.removeClass("loading");
        },
        true
      );
      return false;
    });
  },

  "/settings/userpage": function () {
    var lastTimeout = null;
    $("textarea[name='data']").on("input", function () {
      if (lastTimeout !== null) {
        clearTimeout(lastTimeout);
      }
      var v = $(this).val();
      lastTimeout = setTimeout(function () {
        $("#userpage-content").addClass("loading");
        $.post(
          "/settings/userpage/parse",
          $("textarea[name='data']").val(),
          function (data) {
            $("#userpage-content").removeClass("loading").html(data);
          },
          "text"
        );
      }, 800);
    });
    $("form").submit(function (e) {
      e.preventDefault();
      var obj = formToObject($(this));
      var f = $(this);
      api(
        "users/self/userpage",
        obj,
        function (data) {
          showMessage("success", "Your userpage has been saved.");
          f.removeClass("loading");
        },
        function (data) {
          showMessage("error", data.message);
          f.removeClass("loading");
        },
        true
      );
      return false;
    });
  },

  "/settings/connections": function () {
    $("#unlink-discord").click(function (e) {
      e.preventDefault();
      api(
        "users/self/connections/unlink-discord",
        {},
        function () {
          showMessage("success", "Your discord account has been unlinked.");
          setTimeout(() => { window.location.reload(); }, 1000);
        },
        true
      );
    });
  },

  "/settings/clansettings": function () {
    $("#ginvite").click(function (e) {
      e.preventDefault();
      api(
        "clans/invite",
        {},
        function (data) {
          if (!data.invite) {
            return;
          }
          console.log(data);
          $("#invin").attr("value", data.invite);
          $("#invin").attr("style", "background-color: rgba(0,128,0,.5)");
          setTimeout(() => $("#invin").attr("style", null), 1100);
        },
        true
      );
    });

    $("form#register-form").submit(function (e) {
      e.preventDefault();
      var obj = formToObject($(this));
      var f = $(this);
      return api(
        "clans/settings",
        obj,
        function (data) {
          if (data.message === "tag already exists") {
            showMessage("error", "A clan with that tag already exists!");
            f.removeClass("loading");
            return;
          }

          showMessage("success", "Clan Settings have been saved.");
          f.removeClass("loading");
        },
        true
      );
    });
  },

  "/premium": function () {
    var sl = $("#months-slider")[0];
    noUiSlider.create(sl, {
      start: [1],
      step: 1,
      connect: [true, false],
      range: {
        min: [1],
        max: [36],
      },
    });
    var us = sl.noUiSlider;
    us.on("update", function () {
      var months = us.get();
      var priceUSD = months * PREMIUM_PRICE_PER_MONTH;
      $("#cost").html(
        T("The cost of <b>{{ months }}</b> {{ monthSpelling }} of Akatsuki<sup>+</sup> is <b>${{ amount }} USD</b>", {
          monthSpelling: months == 1 ? T("month") : T("months"),
          months: (+months).toFixed(0),
          amount: priceUSD.toFixed(2),
        })
      );
      $("input[name='os0']").attr(
        "value",
        (+months).toFixed(0) + " month" + (months == 1 ? "" : "s")
      );
      $("#paypal-amt").val(priceUSD.toFixed(2));
    });
  },

  "/settings/avatar": function () {
    $("#file").change(function (e) {
      var f = e.target.files;
      if (f.length < 1) {
        return;
      }
      var u = window.URL.createObjectURL(f[0]);
      var i = $("#avatar-img")[0];
      i.src = u;
      i.onload = function () {
        window.URL.revokeObjectURL(this.src);
      };
    });
  },

  "/settings/profbackground": function () {
    $("#colorpicker").minicolors({
      inline: true,
    });
    $("#background-type").change(function () {
      $("[data-type]:not([hidden])").attr("hidden", "hidden");
      $("[data-type=" + $(this).val() + "]").removeAttr("hidden");
    });
    $("#file").change(function (e) {
      var f = e.target.files;
      if (f.length < 1) {
        return;
      }
      var u = window.URL.createObjectURL(f[0]);
      var i = document.createElement("img");
      i.src = u;
      i.onload = function () {
        window.URL.revokeObjectURL(this.src);
      };
      $("#image-background").empty().append(i);
    });
  },

  "/dev/tokens": function () {
    $("#privileges-number").on("input", function () {
      $("#privileges-text").text(privilegesToString($(this).val()));
    });
  },
};

$(document).ready(function () {
  // semantic stuff
  $(".message .close").on("click", closeClosestMessage);
  $(".ui.checkbox").checkbox();
  $(".ui.dropdown").dropdown();
  $(".ui.progress").progress();
  $(".ui.form").submit(function (e) {
    var t = $(this);
    if (t.hasClass("loading") || t.hasClass("disabled")) {
      e.preventDefault();
      return false;
    }
    t.addClass("loading");
    var f = t.attr("id");
    $("[form='" + f + "']").addClass("loading");
  });

  // amplitude
  const AMPLITUDE_API_KEY = "d24b21f57762f540f5b9c9791b7e3f91";
  amplitude
    .init(AMPLITUDE_API_KEY, { minIdLength: 4 })
    .promise.then(function () {
      const isAuthed = window.currentUserID && window.currentUserID !== "0";
      const hasAmpUserId = !!window.amplitude.getUserId();

      if (isAuthed && !hasAmpUserId) {
        window.amplitude.setUserId(window.currentUserID);
      } else if (!isAuthed && hasAmpUserId) {
        window.amplitude.reset();
      }
    });

  // ripple stuff
  var f = singlePageSnippets[window.location.pathname];
  if (typeof f === "function") f();
  if (typeof deferredToPageLoad === "function") deferredToPageLoad();

  // setup user search
  $("[id=user-search]").search({
    onSelect: function (val) {
      window.location.href = val.url;
      return false;
    },
    apiSettings: {
      url: "/api/v1/users/lookup?name={query}",
      onResponse: function (resp) {
        var r = {
          results: [],
        };
        $.each(resp.users, function (index, item) {
          r.results.push({
            title: item.username,
            url: "/u/" + item.id,
            image: hanayoConf.avatars + "/" + item.id,
          });
        });
        return r;
      },
    },
  });
  $("[id=user-search-input]").keypress(function (e) {
    if (e.which == 13) {
      window.location.pathname = "/u/" + $(this).val();
    }
  });

  $(document).keydown(function (e) {
    var activeElement = $(document.activeElement);
    var isInput = activeElement.is(":input,[contenteditable]");
    if ((e.which === 83 || e.which === 115) && !isInput) {
      $("[id=user-search-input]").focus();
      e.preventDefault();
    }
    if (e.which === 27 && isInput) {
      activeElement.blur();
    }
  });

  // setup timeago
  $.timeago.settings.allowFuture = true;
  $("time.timeago").timeago();

  // Language selector
  $(".language-select").on("click", function (event) {
    event.stopPropagation();
    event.stopImmediatePropagation();
    var lang = $(this).data("lang");
    document.cookie = "language=" + lang + ";path=/;max-age=31536000";
    window.location.reload();
  });

  // Mobile navbar dropdowns
  $("[data-menu]").on("click", function () {
    var dropdownType = $(this).data("menu");
    var isInvoked = $(this).data("invoked") == "true";
    var $dropdown = $('[data-dropdown-menu="' + dropdownType + '"]');
    var $dropdownIcon = $('[data-dropdown-icon="' + dropdownType + '"]');

    if (!isInvoked) {
      $(this).data("invoked", "true");
      $dropdown.slideDown(300, function () {
        $dropdown.css("display", "");
        $dropdownIcon.removeClass("fa-caret-down").addClass("fa-caret-up");
      });
    } else {
      $(this).data("invoked", "false");
      $dropdown.slideUp(300, function () {
        $dropdown.css("display", "none");
        $dropdownIcon.removeClass("fa-caret-up").addClass("fa-caret-down");
      });
    }
  });

  // hook submit button to handle validation of username input
  // on donation pages. This code functions both for the /support
  // as well as the /premium page.
  $("#paypal-form").on("submit", function (e) {
    const le = $(this);
    e.preventDefault();
    api(
      "users/whatid",
      { name: $("#username-input").val() },
      function (data) {
        $("form>input[name='custom']").attr("value", `userid=${data.id}`);
        le.off("submit").trigger("submit");
      },
      function (data) {
        showMessage(
          "error",
          "The username you input does not seem to exist in our systems. " +
            "Please check their username and ensure that it is correct."
        );
      },
      false
    );
  });
});

function closeClosestMessage() {
  $(this)
    .closest(".message")
    .fadeOut(300, function () {
      $(this).remove();
    });
}

function showMessage(type, message) {
  var newEl = $(
    '<div class="ui ' +
      type +
      ' message hidden"><i class="close icon"></i>' +
      T(message) +
      "</div>"
  );
  newEl.find(".close.icon").click(closeClosestMessage);
  $("#messages-container").append(newEl);
  newEl.slideDown(300);
}

function showIdMessage(type, message, id) {
  var newEl = $(
    '<div id="' +
      id +
      '" class="ui ' +
      type +
      ' message hidden"><i class="close icon"></i>' +
      T(message) +
      "</div>"
  );
  newEl.find(".close.icon").click(closeClosestMessage);
  $("#messages-container").append(newEl);
  newEl.slideDown(300);
}

// function for all api calls
function api(endpoint, data, success, failure, post) {
  if (typeof data == "function") {
    success = data;
    data = null;
  }
  if (typeof failure == "boolean") {
    post = failure;
    failure = undefined;
  }

  if (endpoint == "leaderboard") {
    $("#main").addClass("loading");
  } else if (endpoint == "users/achievements") {
    $("#achievements").addClass("loading");
  }

  var errorMessage =
    "An error occurred while contacting the Akatsuki API. Please report this to an Akatsuki developer.";

  $.ajax({
    method: post ? "POST" : "GET",
    dataType: "json",
    url: hanayoConf.baseAPI + "/" + endpoint,
    data: post ? JSON.stringify(data) : data,
    contentType: post ? "application/json; charset=utf-8" : "",
    success: function (data) {
      if (data.code !== undefined && data.code >= 500) {
        console.warn(data);
        showMessage("error", errorMessage);
      }

      if (endpoint == "leaderboard") {
        $("#main").removeClass("loading");
      } else if (endpoint == "users/achievements") {
        $("#achievements").removeClass("loading");
      }

      success(data);
    },
    error: function (jqXHR, textStatus, errorThrown) {
      if (
        jqXHR.status >= 400 &&
        jqXHR.status < 500 &&
        typeof failure == "function"
      ) {
        failure(jqXHR.responseJSON);
        return;
      }
      console.warn(jqXHR, textStatus, errorThrown);
      showMessage("error", errorMessage);
    },
  });
}

var modes = {
  0: "osu!",
  1: "osu!taiko",
  2: "osu!catch",
  3: "osu!mania",
};

var modesShort = {
  0: "std",
  1: "taiko",
  2: "ctb",
  3: "mania",
};

var entityMap = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;",
  '"': "&quot;",
  "'": "&#39;",
  "/": "&#x2F;",
};

function escapeHTML(str) {
  return String(str).replace(/[&<>"'\/]/g, function (s) {
    return entityMap[s];
  });
}

function setupSimplepag(callback) {
  var el = $(".simplepag");
  el.find(".left.floated .item").on("click", function () {
    if ($(this).hasClass("disabled")) return false;
    page--;
    callback();
  });
  el.find(".right.floated .item").on("click", function () {
    if ($(this).hasClass("disabled")) return false;
    page++;
    callback();
  });
}

function disableSimplepagButtons(right) {
  var el = $(".simplepag");

  if (page <= 1) el.find(".left.floated .item").addClass("disabled");
  else el.find(".left.floated .item").removeClass("disabled");

  if (right) el.find(".right.floated .item").addClass("disabled");
  else el.find(".right.floated .item").removeClass("disabled");
}

window.URL = window.URL || window.webkitURL;

// thank mr stackoverflow
function addCommas(nStr) {
  nStr += "";
  x = nStr.split(".");
  x1 = x[0];
  x2 = x.length > 1 ? "." + x[1] : "";
  var rgx = /(\d+)(\d{3})/;
  while (rgx.test(x1)) {
    x1 = x1.replace(rgx, "$1" + "," + "$2");
  }
  return x1 + x2;
}

// helper functions copied from user.js in old-frontend
function getScoreMods(m, noplus) {
  var r = [];
  // has nc => remove dt
  if ((m & 512) == 512) m = m & ~64;
  // has pf => remove sd
  if ((m & 16384) == 16384) m = m & ~32;
  modsString.forEach(function (v, idx) {
    var val = 1 << idx;
    if ((m & val) > 0) r.push(v);
  });
  if (r.length > 0) {
    return (noplus ? "" : "+") + r.join("");
  } else {
    return noplus ? T("None") : "";
  }
}

var modsString = [
  "NF",
  "EZ",
  "TD",
  "HD",
  "HR",
  "SD",
  "DT",
  "RX",
  "HT",
  "NC",
  "FL",
  "AU", // Auto.
  "SO",
  "AP", // Autopilot.
  "PF",
  "K4",
  "K5",
  "K6",
  "K7",
  "K8",
  "FI", // FadeIn
  "RN", // Random
  "LM", // LastMod. Cinema?
  "K9",
  "K0",
  "K1",
  "K3",
  "K2",
  "V2",
];

// time format (seconds -> hh:mm:ss notation)
function timeFormat(t) {
  var h = Math.floor(t / 3600);
  t %= 3600;
  var m = Math.floor(t / 60);
  var s = t % 60;
  var c = "";
  if (h > 0) {
    c += h + ":";
    if (m < 10) {
      c += "0";
    }
    c += m + ":";
  } else {
    c += m + ":";
  }
  if (s < 10) {
    c += "0";
  }
  c += s;
  return c;
}

// http://stackoverflow.com/a/901144/5328069
function query(name, url) {
  if (!url) {
    url = window.location.href;
  }
  name = name.replace(/[\[\]]/g, "\\$&");
  var regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
    results = regex.exec(url);
  if (!results) return null;
  if (!results[2]) return "";
  return decodeURIComponent(results[2].replace(/\+/g, " "));
}

// Useful for forms contacting the Ripple API
function formToObject(form) {
  var inputs = form.find("input, textarea, select");
  var obj = {};
  inputs.each(function (_, el) {
    el = $(el);
    if (el.attr("name") === undefined) {
      return;
    }
    var parts = el.attr("name").split(".");
    var value;
    switch (el.attr("type")) {
      case "checkbox":
        value = el.is(":checked");
        break;
      default:
        switch (el.data("cast")) {
          case "int":
            value = +el.val();
            break;
          default:
            value = el.val();
            break;
        }
        break;
    }
    obj = modifyObjectDynamically(obj, parts, value);
  });
  return obj;
}

// > modifyObjectDynamically({}, ["nice", "meme", "dude"], "lol")
// { nice: { meme: { dude: 'lol' } } }
function modifyObjectDynamically(obj, inds, set) {
  if (inds.length === 1) {
    obj[inds[0]] = set;
  } else if (inds.length > 1) {
    if (typeof obj[inds[0]] !== "object") obj[inds[0]] = {};
    obj[inds[0]] = modifyObjectDynamically(obj[inds[0]], inds.slice(1), set);
  }
  return obj;
}

var langWhitelist = [
  "de",
  "it",
  "ko",
  "es",
  "ru",
  "pl",
  "fr",
  "nl",
  "sv",
  "fi",
  "ro",
  "ko",
  "vi",
];
i18next.use(i18nextXHRBackend).init({
  nsSeparator: false,
  keySeparator: false,
  fallbackLng: false,
  lng: hanayoConf.language,
  whitelist: langWhitelist,
  load: "currentOnly",
  backend: { loadPath: "/static/locale/{{lng}}.json" },
});

var i18nLoaded = $.inArray(hanayoConf.language, langWhitelist) === -1;
i18next.on("loaded", function () {
  i18nLoaded = true;
});

function T(s, settings) {
  if (
    typeof settings !== "undefined" &&
    typeof settings.count !== "undefined" &&
    $.inArray(hanayoConf.language, langWhitelist) === -1 &&
    settings.count !== 1
  )
    s = keyPlurals[s];
  return i18next.t(s, settings);
}

var apiPrivileges = [
  "ReadConfidential",
  "Write",
  "ManageBadges",
  "BetaKeys",
  "ManageSettings",
  "ViewUserAdvanced",
  "ManageUser",
  "ManageRoles",
  "ManageAPIKeys",
  "Blog",
  "APIMeta",
  "Beatmap",
];

function privilegesToString(privs) {
  var privList = [];
  apiPrivileges.forEach(function (value, index) {
    if ((privs & (1 << (index + 1))) != 0) privList.push(value);
  });
  return privList.join(", ");
}

function toggleNavbar() {
  $(".mobile-header").toggleClass("active");
}

function countryToCodepoints(country) {
  chars = [];
  country = country.toUpperCase();

  for (let i = 0; i < country.length; i++) {
    chars.push((country.codePointAt(i) + 127397).toString(16));
  }

  return chars.join("-");
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

function toggleBBCodeBox(e) {
  identifier = $(e).attr("id").split("-")[1];
  contentBox = $("#content-" + identifier);
  icon = $("#icon-" + identifier);

  if (contentBox.hasClass("bbcode-hidden")) {
    contentBox.slideDown(100, () => {
      contentBox.removeClass("bbcode-hidden");
      icon.removeClass("fa-angle-right").addClass("fa-angle-down");
    });
  } else {
    contentBox.slideUp(100, () => {
      contentBox.addClass("bbcode-hidden");
      icon.removeClass("fa-angle-down").addClass("fa-angle-right");
    });
  }
}
