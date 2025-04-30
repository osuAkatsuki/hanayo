// The code quality of it might not be the best, but it surely does the job.
const matchName = $("#match-name");
const matchContainer = $("#match-container");

let intervalId = -1;
let isIntervalNeeded = false;

let firstCurrentEventId = 9999999999999;
let lastCurrentEventId = -1;
let currentGameId = null;

var currentEvents = [];

// Convert enums to values
const teamStyle = {
  0: "team-no-team",
  1: "team-blue",
  2: "team-red",
};

const teamString = {
  0: "noteam",
  1: "blue",
  2: "red",
};

const scoringName = {
  0: "Score",
  1: "Accuracy",
  2: "Combo",
  3: "Score V2",
};

const matchTeamName = {
  0: "Head-to-Head",
  1: "Tag Co-op",
  2: "Team VS",
  3: "Tag Team VS",
};

const osuModeName = {
  0: "osu!",
  1: "Taiko",
  2: "Catch",
  3: "Mania",
};

const scoringTypeField = {
  0: "score",
  1: "accuracy",
  2: "max_combo",
  3: "score",
};

// TODO: is there a way to refactor this nicely?
const eventIcon = {
  MATCH_CREATION: `<div class="event-icon-green"><i class="fas fa-plus"></i></div>`,
  MATCH_USER_JOIN: `<div class="event-icon-green"><i class="fas fa-arrow-right"></i><i class="far fa-circle"></i></div>`,
  MATCH_HOST_ASSIGNMENT: `<div class="event-icon-green"><i class="fas fa-exchange-alt"></i></div>`,
  MATCH_USER_LEFT: `<div class="event-icon-yellow"><i class="fas fa-arrow-left"></i><i class="far fa-circle"></i></div>`,
  MATCH_DISBAND: `<div class="event-icon-yellow"><i class="fas fa-times"></i></div>`,
};

function getScoringOrder(game, score) {
  switch (game.scoring_type) {
    case 1:
      return [
        ["Combo", score.max_combo.toLocaleString()],
        ["Score", score.score.toLocaleString()],
        ["Accuracy", `${score.accuracy.toFixed(2)}%`],
      ];
    case 2:
      return [
        ["Score", score.score.toLocaleString()],
        ["Accuracy", `${score.accuracy.toFixed(2)}%`],
        ["Combo", score.max_combo.toLocaleString()],
      ];
    default:
      return [
        ["Combo", score.max_combo.toLocaleString()],
        ["Accuracy", `${score.accuracy.toFixed(2)}%`],
        ["Score", score.score.toLocaleString()],
      ];
  }
}

function getEventText(event) {
  switch (event.type) {
    case "MATCH_CREATION":
      return `<a href="/u/${event.user.id}">${escapeHTML(event.user.username)}</a> created the match`;
    case "MATCH_USER_JOIN":
      return `<a href="/u/${event.user.id}">${escapeHTML(event.user.username)}</a> joined the match`;
    case "MATCH_HOST_ASSIGNMENT":
      return `<a href="/u/${event.user.id}">${escapeHTML(event.user.username)}</a> became the host`;
    case "MATCH_USER_LEFT":
      return `<a href="/u/${event.user.id}">${escapeHTML(event.user.username)}</a> left the match`;
    case "MATCH_DISBAND":
      return "the match was disbanded";
    default:
      return "unknown event"; // should be unreachable
  }
}

function generateTimeFormat(date) {
  const hours = date.getHours().toString().padStart(2, "0");
  const minutes = date.getMinutes().toString().padStart(2, "0");
  const seconds = date.getSeconds().toString().padStart(2, "0");

  return `${hours}:${minutes}:${seconds}`;
}

function generateFullDateFormat(date) {
  const options = {
    year: "numeric",
    month: "long",
    day: "numeric",
  };

  return `${date.toLocaleDateString("en-GB", options)} ${generateTimeFormat(
    date
  )}`;
}

function buildEvent(event) {
  var eventTime = new Date(event.timestamp);
  return `
    <div class="event-container">
      <span data-tooltip="${generateFullDateFormat(eventTime)}">
        ${generateTimeFormat(eventTime)}
      </span>
      ${eventIcon[event.type]}
      <div>${getEventText(event)}</div>
    </div>
  `;
}

function buildEvents(events) {
  return `
    <div class="ui segment events-segment">
      <div class="events-container">
        ${events.map((event) => buildEvent(event)).join("")}
      </div>
    </div>
  `;
}

function buildScoringOrder(scoring, idx) {
  const attributeLast = idx == 2 ? "attribute-last" : "";
  const scoringValue = idx == 2 ? "scoring-value" : "";

  return `
    <div class="score-box-attribute ${attributeLast}">
      <div class="score-box-attribute-name">${scoring[0]}</div>
      <div class="score-box-attribute-value ${scoringValue}">${scoring[1]}</div>
    </div>
  `;
}

function capitalise(string) {
  return string.charAt(0).toUpperCase() + string.slice(1).toLowerCase();
}

function calculateTeamScoring(game) {
  var scoring = {
    blue: 0,
    red: 0,
  };

  var teamCount = {
    blue: 0,
    red: 0,
  };

  var scoringField = scoringTypeField[game.scoring_type];
  game.scores.forEach((score) => {
    if (!score.passed) return;

    var team = teamString[score.team];
    scoring[team] += score[scoringField];
    teamCount[team]++;
  });

  var averageScoring = game.scoring_type == 1 || game.scoring_type == 2;
  if (averageScoring) {
    scoring.blue /= teamCount.blue;
    scoring.red /= teamCount.red;
  }

  return scoring;
}

function formatToScoring(scoringType, value) {
  switch (scoringType) {
    case 1:
      return `${value.toFixed(2)}%`;
    default:
      return Math.floor(value).toLocaleString();
  }
}

function buildScores(game) {
  var teamMatch = game.team_type == 2 || game.team_type == 3;

  var htmlTemplate = game.scores
    .map((score) => buildScore(game, score))
    .join("");

  if (!teamMatch) {
    return htmlTemplate;
  }

  var scoring = calculateTeamScoring(game);
  var winner = scoring.blue > scoring.red ? "blue" : "red";
  var loser = winner == "blue" ? "red" : "blue";

  var scoreDifference = scoring[winner] - scoring[loser];
  var draw = scoreDifference == 0;

  htmlTemplate += `
    <div class="team-score-displayer">
      <div class="team-score-container">
        <span class="team-name">Blue Team</span>
        <span class="team-score">${formatToScoring(
          game.scoring_type,
          scoring.blue
        )}</span>
      </div>
      <div class="team-score-container">
        <span class="team-name">Red Team</span>
        <span class="team-score">${formatToScoring(
          game.scoring_type,
          scoring.red
        )}</span>
      </div>
    </div>
    <div class="team-difference-win">`;

  if (draw) {
    htmlTemplate += `The game ended in a draw.`;
  } else {
    htmlTemplate += `
      <b>${capitalise(winner)} Team</b> wins by ${formatToScoring(
      game.scoring_type,
      scoreDifference
    )}`;
  }

  htmlTemplate += `</div>`;

  return htmlTemplate;
}

function buildScore(game, score) {
  var countryCodepoints = countryToCodepoints(score.user.country);
  var scoringOrder = getScoringOrder(game, score);
  return `
    <div class="score-box">
        <div class=${teamStyle[score.team]}></div>
        <div class="score-box-details">
            <div class="score-box-details-left">
                <div class="score-box-username">
                    <a href="/u/${score.user.id}">${escapeHTML(score.user.username)}</a>
                </div>
                <a class="score-flag-link" href="/leaderboard?mode=0&rx=0&country=${score.user.country.toLowerCase()}">
                    <img class="score-box-user-flag" src="/static/images/flags/${countryCodepoints}.svg" />
                </a>
            </div>
            <div class="score-box-details-right">
                <div class="score-box-right-top">
                    ${
                      !score.passed
                        ? "<div class='score-box-failed'>FAILED</div>"
                        : ""
                    }
                    <div class="score-box-right-mods">${getScoreMods(
                      score.mods,
                      false
                    )}</div>
                    ${scoringOrder
                      .map((scoring, idx) => buildScoringOrder(scoring, idx))
                      .join("")}
                </div>
                <div class="score-box-right-bottom">
                    <div class="score-box-attribute">
                        <div class="score-box-attribute-name">300</div>
                        <div class="score-box-attribute-value">${score.count_300.toLocaleString()}</div>
                    </div>
                    <div class="score-box-attribute">
                        <div class="score-box-attribute-name">100</div>
                        <div class="score-box-attribute-value">${score.count_100.toLocaleString()}</div>
                    </div>
                    <div class="score-box-attribute">
                        <div class="score-box-attribute-name">50</div>
                        <div class="score-box-attribute-value">${score.count_50.toLocaleString()}</div>
                    </div>
                    <div class="score-box-attribute">
                        <div class="score-box-attribute-name">Miss</div>
                        <div class="score-box-attribute-value">${score.count_miss.toLocaleString()}</div>
                    </div>
                    <div class="score-box-attribute">
                        <div class="score-box-attribute-name">Geki</div>
                        <div class="score-box-attribute-value">${score.count_geki.toLocaleString()}</div>
                    </div>
                    <div class="score-box-attribute attribute-last">
                        <div class="score-box-attribute-name">Katu</div>
                        <div class="score-box-attribute-value">${score.count_katu.toLocaleString()}</div>
                    </div>
                </div>
            </div>
        </div>
    </div>
    `;
}

function buildGame(game) {
  var gameStartTime = new Date(game.start_time);
  var gameEndTime = game.end_time !== null ? new Date(game.end_time) : null;
  return `
    <div class="ui segment game-segment">
        <a class="game-header" href="/b/${game.beatmap.id}">
            <div class="game-image" style="--game-image-url: url('https://assets.ppy.sh/beatmaps/${
              game.beatmap.beatmapset_id
            }/covers/cover.jpg')"></div>
            <div class="game-header-top-text">
                <div class="shadow-text">
                <span data-tooltip="${generateFullDateFormat(gameStartTime)}">
                  ${generateTimeFormat(gameStartTime)}
                </span>
                -
                ${
                  gameEndTime
                    ? `
                    <span data-tooltip="${generateFullDateFormat(gameEndTime)}">
                      ${generateTimeFormat(gameEndTime)}
                    </span>`
                    : "Game in progress"
                }</div>
                <div class="shadow-text">${osuModeName[game.mode]}</div>
                <div class="shadow-text">Highest ${
                  scoringName[game.scoring_type]
                }</div>
            </div>
            <div class="game-header-map-text">
              <div class="game-header-map-title">
                <div class="shadow-text">
                  ${game.beatmap.title} [${game.beatmap.version}]
                </div>
              </div>
              <div class="game-header-map-artist">
                <div class="shadow-text">
                  ${game.beatmap.artist}
                </div>
              </div>
            </div>
            <div class="game-header-mods">
              <div class="shadow-text">
                ${getScoreMods(game.mods, false)}
              </div>
            </div>
            <div class="game-header-team-type">
              <div class="shadow-text">
                ${matchTeamName[game.team_type]}
              </div>
            </div>
        </a>
        <div class="scores-container">
            ${
              !gameEndTime
                ? `
                <div class="ui active centered inline loader"></div>
                <div class="loading-circle-text">Game still in progress</div>
            `
                : !game.scores
                ? `
                    <div class="loading-circle-text">No scores for this game!</div>
                `
                : buildScores(game)
            }
        </div>
    </div>
    `;
}

function renderEvents() {
  // cache match container locally
  var matchContainerJQ = matchContainer;

  var events = [];
  currentEvents.forEach((event) => {
    // treat game playthroughs like a event separator
    if (event.type == "MATCH_GAME_PLAYTHROUGH") {
      if (events.length > 0) {
        matchContainerJQ.append(buildEvents(events));
        events = [];
      }

      if (event.game !== null) {
        matchContainerJQ.append(buildGame(event.game));
        return;
      }
    }
    events.push(event);
  });

  // render any remaining events
  if (events.length > 0) {
    matchContainerJQ.append(buildEvents(events));
  }
}

async function loadMatchData(loadOld = false, loadNew = false) {
  var additionalParams = "";

  if (loadOld) {
    additionalParams = `&before=${firstCurrentEventId}`;
  } else if (loadNew && currentGameId !== null) {
    additionalParams = `&after=${lastCurrentEventId - 1}`;
  } else if (loadNew) {
    additionalParams = `&after=${lastCurrentEventId}`;
  }

  var req = await fetch(
    `/api/v1/match?id=${matchID}&limit=100${additionalParams}`
  );
  var data = await req.json();

  // cache match container locally
  var matchContainerJQ = matchContainer;

  matchContainerJQ.empty();

  const matchNameSafe = escapeHTML(data.match.name);
  matchName.html(`<div class="match-name-div">${matchNameSafe}</div>`);
  document.title = `${matchNameSafe} - Akatsuki`;

  if (data.events !== null) {
    firstCurrentEventId =
      data.events[0].id < firstCurrentEventId
        ? data.events[0].id
        : firstCurrentEventId;

    lastCurrentEventId =
      data.events[data.events.length - 1].id > lastCurrentEventId
        ? data.events[data.events.length - 1].id
        : lastCurrentEventId;

    if (!loadOld && !loadNew) {
      currentEvents.push(...data.events);
    }

    if (loadNew && currentGameId !== null) {
      currentEvents = currentEvents.filter((event) => {
        return event.game?.id !== currentGameId;
      });
    }

    if (loadOld) {
      currentEvents.unshift(...data.events);
    } else if (loadNew) {
      currentEvents.push(...data.events);
    }
  }

  if (data.first_event_id != firstCurrentEventId) {
    matchContainerJQ.append(
      `<div class="ui segment load-old-button-segment">
      <button class="ui labeled icon button" onclick="loadMatchData(true, false)">
        <i class="up arrow icon"></i>
        Show more
      </button>
    </div>`
    );
  }

  currentGameId = data.current_game_id;
  renderEvents();

  // check if we need to start/stop the interval
  var matchEnded = data.match.end_time !== null;

  if (!matchEnded) {
    matchContainerJQ.append(
      `<div class="ui segment game-segment">
        <div class="ui active centered inline loader"></div>
        <div class="loading-circle-text">Match in progress</div>
      </div>`
    );
  }

  if (intervalId !== -1 && matchEnded) {
    clearInterval(intervalId);
  }

  if (intervalId === -1 && !matchEnded) {
    intervalId = setInterval(() => {
      loadMatchData(false, true);
    }, 5000);
  }
}

$(document).ready(() => {
  loadMatchData();
});
