import gulp from "gulp";
import { src, dest, watch, series, parallel } from "gulp";
import babel from "gulp-babel";
import concat from "gulp-concat";
import uglify from "gulp-uglify";
import rename from "gulp-rename";
import cleanCSS from "gulp-clean-css";
import rev from "gulp-rev";
import { deleteAsync } from "del";
import { writeFileSync, existsSync, readFileSync, mkdirSync } from "fs";

var paths = {
  styles: {
    src: "src/css/**/*.css",
    dest: "static/css/",
  },
  scripts: {
    src: "src/js/pages/*.js",
    dest: "static/js/pages/",
  },
  dist: {
    src: [
      "node_modules/jquery/dist/jquery.min.js",
      "node_modules/timeago/jquery.timeago.js",
      "node_modules/i18next/i18next.min.js",
      "node_modules/i18next-xhr-backend/i18nextXHRBackend.min.js",
      "node_modules/fomantic-ui/dist/components/api.min.js",
      "src/js/semantic.min.js",
      "src/js/key_plural.js",
      "src/js/akatsuki_src.js",
    ],
    dest: "static/js/",
  },
};

export const clean = () => deleteAsync(["static/js", "static/css", "static/manifest.json"]);

export function styles() {
  return src(paths.styles.src)
    .pipe(cleanCSS())
    .pipe(rename({ suffix: ".min" }))
    .pipe(rev())
    .pipe(dest(paths.styles.dest))
    .pipe(rev.manifest("css-manifest.json"))
    .pipe(dest("static/"));
}

export function scripts() {
  return src(paths.scripts.src, { sourcemaps: true })
    .pipe(babel())
    .pipe(uglify())
    .pipe(rename({ suffix: ".min" }))
    .pipe(rev())
    .pipe(dest(paths.scripts.dest))
    .pipe(rev.manifest("scripts-manifest.json"))
    .pipe(dest("static/"));
}

export function dist() {
  return src(paths.dist.src, { sourcemaps: true })
    .pipe(babel())
    .pipe(uglify())
    .pipe(concat("dist.min.js"))
    .pipe(rev())
    .pipe(dest(paths.dist.dest))
    .pipe(rev.manifest("dist-manifest.json"))
    .pipe(dest("static/"));
}

// Merge all manifests into one with proper path prefixes
export function mergeManifests(cb) {
  const manifest = {};

  // CSS manifest
  const cssManifestPath = "static/css-manifest.json";
  if (existsSync(cssManifestPath)) {
    const cssManifest = JSON.parse(readFileSync(cssManifestPath, "utf8"));
    for (const [key, value] of Object.entries(cssManifest)) {
      manifest["/static/css/" + key] = "/static/css/" + value;
    }
    deleteAsync([cssManifestPath]);
  }

  // Scripts manifest (page-specific JS)
  const scriptsManifestPath = "static/scripts-manifest.json";
  if (existsSync(scriptsManifestPath)) {
    const scriptsManifest = JSON.parse(readFileSync(scriptsManifestPath, "utf8"));
    for (const [key, value] of Object.entries(scriptsManifest)) {
      manifest["/static/js/pages/" + key] = "/static/js/pages/" + value;
    }
    deleteAsync([scriptsManifestPath]);
  }

  // Dist manifest (main bundle)
  const distManifestPath = "static/dist-manifest.json";
  if (existsSync(distManifestPath)) {
    const distManifest = JSON.parse(readFileSync(distManifestPath, "utf8"));
    for (const [key, value] of Object.entries(distManifest)) {
      manifest["/static/js/" + key] = "/static/js/" + value;
    }
    deleteAsync([distManifestPath]);
  }

  writeFileSync("static/manifest.json", JSON.stringify(manifest, null, 2));
  cb();
}

function watchFiles() {
  watch(paths.scripts.src, scripts);
  watch(paths.dist.src, dist);
  watch(paths.styles.src, styles);
}
export { watchFiles as watch };

const build = series(clean, parallel(styles, scripts, dist), mergeManifests);

export default build;
