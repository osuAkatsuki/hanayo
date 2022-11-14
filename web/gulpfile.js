var gulp    = require("gulp")
var chug    = require("gulp-chug")
var plumber = require("gulp-plumber")
var uglify  = require("gulp-uglify")
var flatten = require("gulp-flatten")
var concat  = require("gulp-concat")
var babel   = require("gulp-babel")
var saveLicense = require('uglify-save-license');

gulp.task("minify-js", async function() {
	gulp
		.src([
			"node_modules/jquery/dist/jquery.min.js",
			"node_modules/timeago/jquery.timeago.js",
			"static/js/semantic.min.js",
			"node_modules/i18next/i18next.min.js",
			"node_modules/i18next-xhr-backend/i18nextXHRBackend.min.js",
			"static/js/key_plural.js",
			"static/js/akatsuki_src.js",
		])
		.pipe(plumber())
		.pipe(concat("dist.min.js"))
		/*.pipe(babel({
			presets: ["latest"]
		})) breaks vue */
		.pipe(flatten())
		.pipe(uglify({
			output: {
				comments: saveLicense
			},
			mangle: true,
		}))
		.pipe(gulp.dest("./static/js"))
})

gulp.task("build", gulp.series("minify-js"))
gulp.task("default", gulp.series("build"))

gulp.task("watch", function() {
	gulp.watch(["static/js/*.js", "!static/js/dist.min.js"], ["minify-js"])
	gulp.watch("semantic/src/**/*", ["build-semantic"])
})

gulp.task("build-semantic", function() {
	gulp.src("./semantic/gulpfile.js")
		.pipe(chug({
			tasks: ['build']
		}))
})
