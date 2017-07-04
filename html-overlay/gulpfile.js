var gulp = require("gulp");
var browserify = require("browserify");
var source = require('vinyl-source-stream');
var tsify = require("tsify");
var watchify = require("watchify");
var gutil = require("gulp-util");

var watchedBrowserify = watchify(browserify({
    basedir: '.',
    debug: true,
    entries: ['src/core.ts'],
    cache: {},
    packageCache: {}
}).plugin(tsify));

function compile() {
    return watchedBrowserify
        .bundle()
        .pipe(source('core.js'))
        .pipe(gulp.dest("www-dist"));
}

gulp.task("default", compile);
watchedBrowserify.on("update", compile);
watchedBrowserify.on("log", gutil.log);
