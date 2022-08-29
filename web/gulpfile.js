var gulp    = require("gulp")
var chug    = require("gulp-chug")
var plumber = require("gulp-plumber")
var uglify  = require("gulp-uglify")
var flatten = require("gulp-flatten")
var concat  = require("gulp-concat")
var babel   = require("gulp-babel")

gulp.task("default", ["build"])
gulp.task("build", [
	"minify-js",
])

gulp.task("watch", function() {
	gulp.watch(["static/*.js", "!static/dist.min.js"], ["minify-js"])
	gulp.watch("semantic/src/**/*", ["build-semantic"])
})

gulp.task("build-semantic", function() {
	gulp.src("./semantic/gulpfile.js")
		.pipe(chug({
			tasks: ['build']
		}))
})

gulp.task("minify-js", function() {
	gulp
		.src([
			"static/licenseheader.js",
			"node_modules/jquery/dist/jquery.min.js",
			"node_modules/timeago/jquery.timeago.js",
			"static/semantic.min.js",
			"node_modules/i18next/i18next.min.js",
			"node_modules/i18next-xhr-backend/i18nextXHRBackend.min.js",
			"static/key_plural.js",
			"static/ripple.js",
		])
		.pipe(plumber())
		.pipe(concat("dist.min.js"))
		/*.pipe(babel({
			presets: ["latest"]
		})) breaks vue */
		.pipe(flatten())
		.pipe(uglify({
			mangle: true,
			preserveComments: "license"
		}))
		.pipe(gulp.dest("./static"))
})
