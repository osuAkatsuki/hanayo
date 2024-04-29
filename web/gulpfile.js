import gulp from "gulp";
import { src, dest, watch } from "gulp";
import babel from "gulp-babel";
import concat from "gulp-concat";
import uglify from "gulp-uglify";
import rename from "gulp-rename";
import cleanCSS from "gulp-clean-css";
import { deleteAsync } from "del";
import pkg from "gulp-typescript";
const { createProject } = pkg;

const paths = {
  styles: {
    // TODO: add src for semantic-ui
    //       once we move all changes to akatsuki.css
    src: "src/css/**/*.css",
    dest: "static/css/",
  },
  scripts: {
    src: "src/js/pages/*.js",
    dest: "static/js/pages/",
  },
  dist: {
    src: [
      // TODO: add src for semantic-ui
      "node_modules/jquery/dist/jquery.min.js",
      "node_modules/timeago/jquery.timeago.js",
      "node_modules/i18next/i18next.min.js",
      "node_modules/i18next-xhr-backend/i18nextXHRBackend.min.js",
      "node_modules/fomantic-ui/dist/components/api.min.js",
      // "src/js/semantic.min.js",
      "src/js/key_plural.js",
      "src/js/akatsuki_src.js",
    ],
    dest: "static/js/",
  },
};

export const clean = () => deleteAsync(["static/js", "static/css"]);

export function styles() {
  return src(paths.styles.src)
    .pipe(cleanCSS())
    .pipe(
      rename({
        suffix: ".min",
      })
    )
    .pipe(dest(paths.styles.dest));
}

export function scripts() {
  const tsProject = createProject("tsconfig.json");
  return tsProject
    .src()
    .pipe(tsProject())
    .js.pipe(babel())
    .pipe(uglify())
    .pipe(
      rename({
        suffix: ".min",
      })
    )
    .pipe(dest(paths.scripts.dest));
}

export function dist() {
  const tsProject = createProject("tsconfig.json");
  return tsProject
    .src()
    .pipe(tsProject())
    .js.pipe(babel())
    .pipe(uglify())
    .pipe(concat("dist.min.js"))
    .pipe(dest(paths.dist.dest));
}

function watchFiles() {
  watch(paths.scripts.src, scripts);
  watch(paths.dist.src, dist);
  watch(paths.styles.src, styles);
}
export { watchFiles as watch };

const build = gulp.series(clean, gulp.parallel(styles, scripts, dist));

export default build;
