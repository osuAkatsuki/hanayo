// code that is executed on every user profile
$(document).ready(function () {
	var wl = window.location;
	var newPathName = wl.pathname;
	// userID is defined in profile.html
	if (newPathName.split("/")[2] != userID) {
		newPathName = "/u/" + userID;
	}

	let newSearch = wl.search;

	if (wl.search.indexOf("mode=") === -1) {
		newSearch = "?mode=" + favouriteMode;
	}

	if (wl.search.indexOf("rx=") === -1 || wl.search.indexOf("?rx=") != -1)
		newSearch += "&rx=" + preferRelax;

	if (wl.search != newSearch)
		window.history.replaceState('', document.title, newPathName + newSearch + wl.hash);
	else if (wl.pathname != newPathName)
		window.history.replaceState('', document.title, newPathName + wl.search + wl.hash);

	setDefaultScoreTable();
	checkRelaxMania(favouriteMode, preferRelax);

	$("#rx-menu>.simple-banner-swtich").click(function (e) {
		e.preventDefault();
		if ($(this).hasClass("active"))
			return;

		preferRelax = $(this).data("rx");
		checkRelaxMania(favouriteMode, preferRelax);
		$("[data-mode]:not(.simple-banner-swtich):not([hidden])").attr("hidden", "");
		$("[data-mode=" + favouriteMode + "][data-rx=" + preferRelax + "]:not(.simple-banner-swtich)").removeAttr("hidden");
		$("#rx-menu>.active.simple-banner-swtich").removeClass("active");
		var needsLoad = $("#scores-zone>[data-mode=" + favouriteMode + "][data-loaded=0][data-rx=" + preferRelax + "]");
		if (needsLoad.length > 0)
			initialiseScores(needsLoad, favouriteMode);
		$(this).addClass("active");
		window.history.replaceState('', document.title, `${wl.pathname}?mode=${favouriteMode}&rx=${preferRelax}${wl.hash}`)
	});

	// when an item in the mode menu is clicked, it means we should change the mode.
	$("#mode-menu>.simple-banner-swtich").click(function (e) {
		e.preventDefault();
		if ($(this).hasClass("active"))
			return;
		var m = $(this).data("mode");
		favouriteMode = m;
		checkRelaxMania(m, preferRelax);
		$("[data-mode]:not(.simple-banner-swtich):not([hidden])").attr("hidden", "");
		$("[data-mode=" + m + "][data-rx=" + preferRelax + "]:not(.simple-banner-swtich)").removeAttr("hidden");
		$("#mode-menu>.active.simple-banner-swtich").removeClass("active");
		var needsLoad = $("#scores-zone>[data-mode=" + m + "][data-loaded=0][data-rx=" + preferRelax + "]");
		if (needsLoad.length > 0)
			initialiseScores(needsLoad, m);
		$(this).addClass("active");
		window.history.replaceState('', document.title, `${wl.pathname}?mode=${m}&rx=${preferRelax}${wl.hash}`);
	});
	initialiseAchievements();
	initialiseUserpage();
	initialiseFriends();
	// load scores page for the current favourite mode
	var i = function () { initialiseScores($("#scores-zone>div[data-mode=" + favouriteMode + "][data-rx=" + preferRelax + "]"), favouriteMode) };
	if (i18nLoaded)
		i();
	else
		i18next.on("loaded", function () {
			i();
		});
});

function initialiseUserpage() {
	api("users/userpage", { id: userID }, (resp) => {
		var userpage = $("#userpage-content")

		if (!resp.userpage_compiled) return

		userpage.css("display", "")
		userpage.html(resp.userpage_compiled)
		userpage.removeClass("loading")
	})
}

function initialiseAchievements() {
	api('users/achievements' + (currentUserID == userID ? '?all' : ''),
		{ id: userID }, function (resp) {
			var achievements = resp.achievements;
			// no achievements -- show default message
			if (achievements.length === 0) {
				$("#achievements")
					.append($("<div class='ui sixteen wide column'>")
						.text(T("Nothing here. Yet.")));
				$("#load-more-achievements").remove();
				return;
			}

			var displayAchievements = function (limit, achievedOnly) {
				var $ach = $("#achievements").empty();
				limit = limit < 0 ? achievements.length : limit;
				var shown = 0;
				for (var i = 0; i < achievements.length; i++) {
					var ach = achievements[i];
					if (shown >= limit || (achievedOnly && !ach.achieved)) {
						continue;
					}
					shown++;
					$ach.append(
						$("<div class='ui two wide column'>").append(
							$("<img src='https://s.ripple.moe/images/medals-" +
								"client/" + ach.icon + ".png' alt='" + ach.name +
								"' class='" +
								(!ach.achieved ? "locked-achievement" : "achievement") +
								"'>").popup({
									title: ach.name,
									content: ach.description,
									position: "bottom center",
									distanceAway: 10
								})
						)
					);
				}
				// if we've shown nothing, and achievedOnly is enabled, try again
				// this time disabling it.
				if (shown == 0 && achievedOnly) {
					displayAchievements(limit, false);
				}
			};

			// only 8 achievements - we can remove the button completely, because
			// it won't be used (no more achievements).
			// otherwise, we simply remove the disabled class and add the click handler
			// to activate it.
			if (achievements.length <= 8) {
				$("#load-more-achievements").remove();
			} else {
				$("#load-more-achievements")
					.removeClass("disabled")
					.click(function () {
						$(this).remove();
						displayAchievements(-1, false);
					});
			}
			displayAchievements(8, true);
		});
}

function initialiseFriends() {
	var b = $("#add-friend-button");
	if (b.length == 0) return;
	api('friends/with', { id: userID }, setFriendOnResponse);
	b.click(friendClick);
}
function setFriendOnResponse(r) {
	var x = 0;
	if (r.friend) x++;
	if (r.mutual) x++;
	setFriend(x);
}
function setFriend(i) {
	var b = $("#add-friend-button");
	b.removeClass("loading green blue red");
	switch (i) {
		case 0:
			b
				.addClass("blue")
				.attr("title", T("Add friend"))
				.html(`<i class="fas fa-user-plus"></i>`);
			break;
		case 1:
			b
				.addClass("green")
				.attr("title", T("Remove friend"))
				.html(`<i class="fas fa-user-times"></i>`);
			break;
		case 2:
			b
				.addClass("red")
				.attr("title", T("Unmutual friend"))
				.html(`<i class="fas fa-user-friends"></i>`);
			break;
	}
	b.attr("data-friends", i > 0 ? 1 : 0)
}
function friendClick() {
	var t = $(this);
	if (t.hasClass("loading")) return;
	t.addClass("loading");
	api("friends/" + (t.attr("data-friends") == 1 ? "del" : "add"), { user: userID }, setFriendOnResponse, true);
}

var defaultScoreTable;
function setDefaultScoreTable() {
	defaultScoreTable = $("<div class='score-data' />")
		.append(
			$("<div class='scores' />")
		)
		.append(
			$("<div class='extra-block' />").append(
				$("<a class='show-button'>" + T("Load more") + "</a>").click(loadMoreClick)
			)
		)
		;
}

i18next.on('loaded', function (loaded) {
	setDefaultScoreTable();
});
function initialiseScores(el, mode) {
	el.attr("data-loaded", "1");
	var pinned = defaultScoreTable.clone(true);
	var best = defaultScoreTable.clone(true);

	var most_played = defaultScoreTable.clone(true);

	var first = defaultScoreTable.clone(true);
	var recent = defaultScoreTable.clone(true);

	let rxAsString = preferRelax != 0 ? (preferRelax == 1 ? "r" : "a") : "v"
	let firstSuffix = `${rxAsString}${mode}`

	pinned.attr("data-type", "pinned");
	best.attr("data-type", "best");
	most_played.attr("data-type", "most_played");
	first.attr("data-type", "first");
	recent.attr("data-type", "recent");
	el.append($("<div class='ui segments' />").append(
		$("<div class='ui segment margin sui' />").append(`<div class='header-top'><h2 class='ui header'>${T("Pinned scores")}</h2></div>`, pinned),
		$("<div class='ui segment margin sui' />").append(`<div class='header-top'><h2 class='ui header'>${T("Best scores")}</h2></div>`, best),
		$("<div class='ui segment margin sui' />").append(`<div class='header-top'><h2 class='ui header'>${T("Most played beatmaps")}</h2></div>`, most_played),
		$("<div class='ui segment margin sui' />").append(`<div class='header-top'><h2 class='ui header'>${T("First Place Ranks")} <span id='first-${firstSuffix}' style='font-size: medium;'>(.. in total)</span></h2></div>`, first),
		$("<div class='ui segment margin sui' />").append(`<div class='header-top'><h2 class='ui header'>${T("Recent scores (24h)")}</h2></div>`, recent),
	));
	loadScoresPage("pinned", mode);
	loadScoresPage("best", mode);
	loadMostPlayedBeatmaps("most_played", mode);
	loadScoresPage("first", mode);
	loadScoresPage("recent", mode);
	$('#user-scores').removeClass('load-data')
};
function loadMoreClick() {
	var t = $(this);
	if (t.hasClass("disabled"))
		return;
	t.addClass("disabled");
	var type = t.parents("div[data-type]").data("type");
	var mode = t.parents("div[data-mode]").data("mode");
	loadScoresPage(type, mode);
}

// currentPage for each mode
var currentPage = {
	0: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 },
	1: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 },
	2: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 },
	3: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 }
};

var rPage = {
	0: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 },
	1: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 },
	2: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 },
	3: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 }
};

var aPage = {
	0: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 },
	1: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 },
	2: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 },
	3: { pinned: 0, best: 0, most_played: 0, recent: 0, first: 0 }
};

const scoreNotFoundElement = `<div class="map-single" id="not-found-container">
<div class="map-content1">
	<div class="map-data">
		<div class="map-image" style="background:linear-gradient(rgb(0 0 0 / 70%), rgb(0 0 0 / 70%)); background-size: cover;">
			<div class="map-grade rank-SHD">: (</div>
		</div>
		<div class="map-title-block map-padding">
			<div class="map-title">
				<h4>No score available</h4>
			</div>
			<div class="play-stats">
				Maybe you should play something?
			</div>
		</div>
	</div>
</div>
</div>`

function loadMostPlayedBeatmaps(type, mode) {
	var mostPlayedTable = $("#scores-zone div[data-mode=" + mode + "][data-rx=" + preferRelax + "] div[data-type=" + type + "] .scores");

	var page;
	if (preferRelax == 1) page = ++rPage[mode][type];
	else if (preferRelax == 2) page = ++aPage[mode][type];
	else page = ++currentPage[mode][type];

	api('users/most_played', { id: userID, mode: mode, p: page, l: 5, rx: preferRelax }, function (resp) {
		if (resp.most_played_beatmaps == null) {
			mostPlayedTable.html(scoreNotFoundElement)
			disableLoadMoreButton(type, mode);
			return;
		}

		resp.most_played_beatmaps.forEach(function (el, idx) {
			mostPlayedTable.append(`
			<div class="new map-single" style="cursor: auto">
				<div class="map-content1">
					<div class="map-data">
						<div class="map-image" style="background:linear-gradient( rgb(0 0 0 / 70%), rgb(0 0 0 / 70%) ), url(https://assets.ppy.sh/beatmaps/${el.beatmap.beatmapset_id}/covers/cover@2x.jpg); background-size: cover;">
						</div>
						<div class="map-title-block">
							<div class="map-title">
								<a class="beatmap-link" href="/b/${el.beatmap.beatmap_id}">
									${escapeHTML(el.beatmap.song_name)}
								</a>
							</div>
						</div>
					</div>
				</div>
				<div class="map-content2">
					<div class="score-details d-flex">
						<div class="score-details_right-block">
							<div class="score-details_pp-block">
								<div class="map-pp">
									<i class="fa-solid fa-play"></i> <b>${el.playcount}</b>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>`
			)
		})

		var enable = true;
		if (resp.most_played_beatmaps.length != 5)
			enable = false;

		disableLoadMoreButton(type, mode, enable);
	})
}

var scoreStore = {};
function loadScoresPage(type, mode) {
	var table = $("#scores-zone div[data-mode=" + mode + "][data-rx=" + preferRelax + "] div[data-type=" + type + "] .scores");

	// redirect it to most played load.
	if (type == "most_played") {
		return loadMostPlayedBeatmaps(type, mode)
	}

	var page;
	if (preferRelax == 1) page = ++rPage[mode][type];
	else if (preferRelax == 2) page = ++aPage[mode][type];
	else page = ++currentPage[mode][type];

	if (type != "pinned") {
		api("users/scores/" + type, {
			mode: mode,
			p: page,
			l: 10,
			rx: preferRelax,
			id: userID,
			uid: userID,
			actual_id: window.actualID
		}, function (r) {
			if (type === 'first') {
				let rxAsString = preferRelax != 0 ? (preferRelax == 1 ? "r" : "a") : "v"
				let firstSuffix = `${rxAsString}${mode}`
				$(`#first-${firstSuffix}`).text(`(${r.total} in total)`)
			}

			console.log(r, type)
			if (r.scores == null) {
				disableLoadMoreButton(type, mode);
				table.html(scoreNotFoundElement)
				return;
			} else {
				if (r.scores.length === 0) {
					disableLoadMoreButton(type, mode);
					table.html(scoreNotFoundElement)
					return;
				}
			}

			r.scores.forEach(function (v, idx) {
				scoreStore[v.id] = v;

				if (v.completed != 0) {
					var scoreRank = getRank(mode, v.mods, v.accuracy, v.count_300, v.count_100, v.count_50, v.count_miss);
				} else {
					var scoreRank = 'F';
				}

				table.append(`
				<div class="new map-single complete-${v.completed}" data-scoreid="${v.id}">
					<div class="map-content1">
						<div class="map-data">
							<div class="map-image" style="background:linear-gradient( rgb(0 0 0 / 70%), rgb(0 0 0 / 70%) ), url(https://assets.ppy.sh/beatmaps/${v.beatmap.beatmapset_id}/covers/cover@2x.jpg); background-size: cover;">
								<div class="map-grade rank-${scoreRank}">${scoreRank.replace("HD", "")}</div>
							</div>
							<div class="map-title-block">
								<div class="map-title"><a class="beatmap-link">
									${escapeHTML(v.beatmap.song_name)}
									</a>
								</div>
								<div class="play-stats">
									${addCommas(v.score)} / ${addCommas(v.max_combo)}x / <b>${getScoreMods(v.mods, true)}</b>
								</div>
								<div class="map-date">
									<time class="new timeago" datetime="${v.time}">
										${v.time}
									</time>
								</div>
							</div>
						</div>
					</div>
					<div class="map-content2">
						<div class="score-details d-flex">
							<div class="score-details_right-block">
								<div class="score-details_pp-block">
									<div class="map-pp">
										${ppOrScore(v.pp, v.score)}
									</div>
									<div class="map-acc">accuracy:&nbsp;<b>
										${v.accuracy.toFixed(2)}%
										</b>
									</div>
								</div>
								<div data-btns-score-id='${v.id}'>
									${downloadStar(v.id)}
									${userID == window.actualID && !v.pinned ? pinButton(v.id, preferRelax) : ""}
								</div>
								<div class="score-details_icon-block">
									<i class="angle right icon"></i>
								</div>
							</div>
						</div>
					</div>
				</div>
				`);
			});
			$(".new.timeago").timeago().removeClass("new");
			$(".new.map-single").click(viewScoreInfo).removeClass("new");
			$(".new.downloadstar").click(function (e) {
				e.stopPropagation();
			}).removeClass("new");
			$(".new.pinbutton").click(function (e) {
				e.stopPropagation();
			}).removeClass("new");
			$(".new.unpinbutton").click(function (e) {
				e.stopPropagation();
			}).removeClass("new");
			var enable = true;
			if (r.scores.length != 10)
				enable = false;
			disableLoadMoreButton(type, mode, enable);
		});
	} else {
		pin_api("pinned", {
			mode: mode,
			p: page,
			l: 10,
			rx: preferRelax,
			id: userID,
		}, function (r) {
			if (r.scores.length === 0) {
				table.html(scoreNotFoundElement)
				disableLoadMoreButton(type, mode);
				return;
			}

			r.scores.forEach(score => do_pin(table, score, mode));
			$(".new.timeago").timeago().removeClass("new");
			$(".new.map-single").click(viewScoreInfo).removeClass("new");
			$(".new.downloadstar").click(function (e) {
				e.stopPropagation();
			}).removeClass("new");
			$(".new.pinbutton").click(function (e) {
				e.stopPropagation();
			}).removeClass("new");
			$(".new.unpinbutton").click(function (e) {
				e.stopPropagation();
			}).removeClass("new");
			var enable = true;
			if (r.scores.length != 10)
				enable = false;
			disableLoadMoreButton(type, mode, enable);
		});
	}
}
function refreshTable(type) {
	loadScoresPage(type, mode);
}

function do_pin(table, score, mode) {
	scoreStore[score.id] = score;
	if (score.completed != 0) {
		var scoreRank = getRank(mode, score.mods, score.accuracy, score.count_300, score.count_100, score.count_50, score.count_miss);
	} else {
		var scoreRank = 'F';
	}
	table.append(`
	<div class="new map-single complete-${score.completed}" data-pinnedscoreid="${score.id}">
		<div class="map-content1">
			<div class="map-data">
				<div class="map-image" style="background:linear-gradient( rgb(0 0 0 / 70%), rgb(0 0 0 / 70%) ), url(https://assets.ppy.sh/beatmaps/${score.beatmap.beatmapset_id}/covers/cover@2x.jpg); background-size: cover;">
					<div class="map-grade rank-${scoreRank}">${scoreRank.replace("HD", "")}</div>
				</div>
				<div class="map-title-block">
					<div class="map-title"><a class="beatmap-link">
						${escapeHTML(score.beatmap.song_name)}
						</a>
					</div>
					<div class="play-stats">
						${addCommas(score.score)} / ${addCommas(score.max_combo)}x / <b>${getScoreMods(score.mods, true)}</b>
					</div>
					<div class="map-date">
						<time class="new timeago" datetime="${score.time}">
							${score.time}
						</time>
					</div>
				</div>
			</div>
		</div>
		<div class="map-content2">
			<div class="score-details d-flex">
				<div class="score-details_right-block">
					<div class="score-details_pp-block">
						<div class="map-pp">
							${ppOrScore(score.pp, score.score)}
						</div>
						<div class="map-acc">accuracy:&nbsp;<b>
							${score.accuracy.toFixed(2)}%
							</b>
						</div>
					</div>
					${downloadStar(score.id)}
					${userID == window.actualID ? unpinButton(score.id, preferRelax) : ""}
					<div class="score-details_icon-block">
						<i class="angle right icon"></i>
					</div>
				</div>
			</div>
		</div>
	</div>
	`);
}

function pinSuccess(data) {
	score = scoreStore[data['score_id']]

	var table = $("#scores-zone div[data-mode=" + favouriteMode + "][data-rx=" + preferRelax + "] div[data-type=pinned] .scores");
	if (!table) return showMessage("error", "Tell Flame to fix this");

	notFoundElement = table.find("#not-found-container")
	if (table[0].childElementCount === 1 && notFoundElement) {
		notFoundElement.remove()
	}

	do_pin(table, score, favouriteMode);
	$(".new.timeago").timeago().removeClass("new");
	$(".new.downloadstar").click(function (e) {
		e.stopPropagation();
	}).removeClass("new");
	$(".new.unpinbutton").click(function (e) {
		e.stopPropagation();
	}).removeClass("new");

	var otherScores = $(`[data-btns-score-id="${data['score_id']}"]`)
	otherScores.find('.pinbutton').remove();

	showMessage("success", "Score pinned.");
}

function unpinSuccess(data) {
	var table = $("#scores-zone div[data-mode=" + favouriteMode + "][data-rx=" + preferRelax + "] div[data-type=pinned] .scores");
	var row = $(`div[data-pinnedscoreid=${data['score_id']}]`)
	row.remove()

	if (table[0].childElementCount === 0) {
		table.html(scoreNotFoundElement)
	}

	var otherScores = $(`[data-btns-score-id="${data['score_id']}"]`)
	otherScores.append(pinButton(data['score_id'], preferRelax))

	$(".new.pinbutton").click(function (e) {
		e.stopPropagation();
	}).removeClass("new");

	showMessage("success", "Score unpinned.");
}

function pin_api(endpoint, data, success, failure, post) {
	if (typeof data == "function") {
		success = data;
		data = null;
	}
	if (typeof failure == "boolean") {
		post = failure;
		failure = undefined;
	}

	var errorMessage =
		"An error occurred while contacting the Akatsuki API. Please report this to an Akatsuki developer.";

	$.ajax({
		method: (post ? "POST" : "GET"),
		dataType: "json",
		url: hanayoConf.baseAPI + "/pinned/" + endpoint,
		data: (post ? JSON.stringify(data) : data),
		contentType: (post ? "application/json; charset=utf-8" : ""),
		success: function (data) {
			if (data.code >= 500) {
				console.warn(data);
				showMessage("error", errorMessage);
			}
			success(data);
		},
		error: function (jqXHR, textStatus, errorThrown) {
			if ((jqXHR.status >= 400 && jqXHR.status < 500) &&
				typeof failure == "function") {
				failure(jqXHR.responseJSON);
				return;
			}
			console.warn(jqXHR, textStatus, errorThrown);
			showMessage("error", errorMessage);
		},
	});
};

function pinScore(id, rx) {
	pin_api("pin", { id: id, rx: rx }, pinSuccess, function (data) { }, true);
}

function unpinScore(id, rx) {
	pin_api("unpin", { id: id, rx: rx }, unpinSuccess, function (data) { }, true);
}

function pinButton(id, rx) {
	return `<a href="#" class="new pinbutton" title="Pin Score" onclick='pinScore("${id}", ${rx})'><i class='pin icon'></i></a>`
}

function unpinButton(id, rx) {
	return `<a href="#" class="new unpinbutton" title="Unpin Score" onclick='unpinScore("${id}", ${rx})'><i class='pin icon'></i></a>`
}

function downloadStar(id) {
	return "<a href='/web/replays/" + id + "' title='Download Replay' class='new downloadstar'><i class='fa-solid fa-download icon'></i></a>";
}

function weightedPP(type, page, idx, pp) {
	if (type != "best" || pp == 0)
		return "";
	var perc = Math.pow(0.95, ((page - 1) * 10) + idx);
	var wpp = pp * perc;
	return "<i title='Weighted PP, " + Math.round(perc * 100) + "%'>(" + wpp.toFixed(2) + "pp)</i>";
}
function disableLoadMoreButton(type, mode, enable) {
	var button = $("#scores-zone div[data-mode=" + mode + "][data-rx=" + preferRelax + "] div[data-type=" + type + "] .show-button");
	if (enable) button.removeClass("disabled");
	else button.addClass("disabled");
}
function viewScoreInfo() {
	var scoreid = $(this).data("scoreid");
	if (!scoreid && scoreid !== 0) {
		scoreid = $(this).data("pinnedscoreid");

		if (!scoreid && scoreid !== 0) return;
	}

	var s = scoreStore[scoreid];
	if (s === undefined) return;

	// data to be displayed in the table.
	var data = {
		"Points": addCommas(s.score),
		"PP": addCommas(s.pp),
		"Beatmap": "<a href='/b/" + s.beatmap.beatmap_id + "'>" + escapeHTML(s.beatmap.song_name) + "</a>",
		"Accuracy": s.accuracy + "%",
		"Max combo": addCommas(s.max_combo) + "/" + addCommas(s.beatmap.max_combo)
			+ (s.full_combo ? " " + T("(full combo)") : ""),
		"Difficulty": T("{{ stars }} star", {
			stars: s.beatmap.difficulty2[modesShort[s.play_mode]],
			count: Math.round(s.beatmap.difficulty2[modesShort[s.play_mode]]),
		}),
		"Mods": getScoreMods(s.mods, true),
	};

	// hits data
	var hd = {};
	var trans = modeTranslations[s.play_mode];
	[
		s.count_300,
		s.count_100,
		s.count_50,
		s.count_geki,
		s.count_katu,
		s.count_miss,
	].forEach(function (val, i) {
		hd[trans[i]] = val;
	});

	data = $.extend(data, hd, {
		"Ranked?": T(s.completed == 3 ? "Yes" : "No"),
		"Achieved": s.time,
		"Mode": modes[s.play_mode],
		"File": "<a href='/web/replays/" + s.id + "' class='new downloadstar'>Replay</a>",
	});

	var els = [];
	$.each(data, function (key, value) {
		els.push(
			$("<tr />").append(
				$("<td>" + T(key) + "</td>"),
				$("<td>" + value + "</td>")
			)
		);
	});

	$("#score-data-table tr").remove();
	$("#score-data-table").append(els);
	$(".ui.modal").modal("show");
}

var modeTranslations = [
	[
		"300s",
		"100s",
		"50s",
		"Gekis",
		"Katus",
		"Misses"
	],
	[
		"GREATs",
		"GOODs",
		"50s",
		"GREATs (Gekis)",
		"GOODs (Katus)",
		"Misses"
	],
	[
		"Fruits (300s)",
		"Ticks (100s)",
		"Droplets",
		"Gekis",
		"Droplet misses",
		"Misses"
	],
	[
		"300s",
		"200s",
		"50s",
		"Max 300s",
		"100s",
		"Misses"
	]
];

function getRank(gameMode, mods, acc, c300, c100, c50, cmiss) {
	var total = c300 + c100 + c50 + cmiss;

	// Hidden | Flashlight | FadeIn
	var hdfl = (mods & (1049608)) > 0;

	var ss = hdfl ? "SSHD" : "SS";
	var s = hdfl ? "SHD" : "S";

	switch (gameMode) {
		case 0:
		case 1:
			var ratio300 = c300 / total;
			var ratio50 = c50 / total;

			if (ratio300 == 1)
				return ss;

			if (ratio300 > 0.9 && ratio50 <= 0.01 && cmiss == 0)
				return s;

			if ((ratio300 > 0.8 && cmiss == 0) || (ratio300 > 0.9))
				return "A";

			if ((ratio300 > 0.7 && cmiss == 0) || (ratio300 > 0.8))
				return "B";

			if (ratio300 > 0.6)
				return "C";

			return "D";

		case 2:
			if (acc == 100)
				return ss;

			if (acc > 98)
				return s;

			if (acc > 94)
				return "A";

			if (acc > 90)
				return "B";

			if (acc > 85)
				return "C";

			return "D";

		case 3:
			if (acc == 100)
				return ss;

			if (acc > 95)
				return s;

			if (acc > 90)
				return "A";

			if (acc > 80)
				return "B";

			if (acc > 70)
				return "C";

			return "D";
	}
}

function ppOrScore(pp, score) {
	if (pp != 0)
		return addCommas(Math.round(pp)) + "pp";
	return addCommas(score);
}

function beatmapLink(type, id) {
	if (type == "s")
		return "<a href='/s/" + id + "'>" + id + '</a>';
	return "<a href='/b/" + id + "'>" + id + '</a>';
}

function checkRelaxMania(mode, rx) {
	if (rx == 1) {
		for (i = 0; i <= 3; i++) {
			$(`.simple-banner-swtich[data-mode='${i}']`).removeClass('disabled')
		}
		$(".simple-banner-swtich[data-mode='3']").addClass('disabled')
	} else if (rx == 2) {
		for (i = 1; i <= 3; i++) {
			$(`.simple-banner-swtich[data-mode='${i}']`).addClass('disabled')
		}
	} else {
		for (i = 0; i <= 3; i++) {
			$(`.simple-banner-swtich[data-mode='${i}']`).removeClass('disabled')
		}
	}

	if (mode === 3) {
		$(".simple-banner-swtich[data-rx='1']").addClass('disabled')
		$(".simple-banner-swtich[data-rx='2']").addClass('disabled')
	} else if (mode === 1 || mode === 2) {
		$(".simple-banner-swtich[data-rx='2']").addClass('disabled')
	} else {
		$(".simple-banner-swtich[data-rx='1']").removeClass('disabled')
		$(".simple-banner-swtich[data-rx='2']").removeClass('disabled')
	}
}