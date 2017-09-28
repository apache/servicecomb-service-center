var fs            = require('fs');
var connect       = require('gulp-connect');
var gulp          = require('gulp');
var KarmaServer   = require('karma').Server;
var concat        = require('gulp-concat');
var jshint        = require('gulp-jshint');
var header        = require('gulp-header');
var rename        = require('gulp-rename');
var es            = require('event-stream');
var del           = require('del');
var uglify        = require('gulp-uglify');
var minifyHtml    = require('gulp-minify-html');
var minifyCSS     = require('gulp-minify-css');
var templateCache = require('gulp-angular-templatecache');
var gutil         = require('gulp-util');
var plumber       = require('gulp-plumber');
var open          = require('gulp-open');
var less          = require('gulp-less');
var order         = require("gulp-order");
var runSequence   = require('run-sequence');


var config = {
  pkg : JSON.parse(fs.readFileSync('./package.json')),
  banner:
      '/*!\n' +
      ' * <%= pkg.name %>\n' +
      ' * <%= pkg.homepage %>\n' +
      ' * Version: <%= pkg.version %> - <%= timestamp %>\n' +
      ' * License: <%= pkg.license %>\n' +
      ' */\n\n\n'
};

gulp.task('connect', function() {
  connect.server({
    root: '.',
    livereload: true
  });
});

gulp.task('html', function () {
  gulp.src(['./demo/*.html', '.src/*.html'])
    .pipe(connect.reload());
});

gulp.task('watch', function () {
  gulp.watch(['./demo/**/*.html'], ['html']);
  gulp.watch(['./**/*.less'], ['styles']);
  gulp.watch(['./**/*.js', './**/*.html'], ['scripts']);
});

gulp.task('clean', function(cb) {
  del(['dist'], cb);
});

gulp.task('scripts', function() {

  function buildTemplates() {
    return gulp.src('src/**/*.html')
      .pipe(minifyHtml({
             empty: true,
             spare: true,
             quotes: true
            }))
      .pipe(templateCache({module: 'jsonFormatter'}));
  };

  function buildDistJS(){
    return gulp.src([
      'src/json-formatter.js',
      'src/recursion-helper.js'
      ])
      .pipe(concat('json-formatter.js'))
      .pipe(plumber({
        errorHandler: handleError
      }))
      .pipe(jshint())
      .pipe(jshint.reporter('jshint-stylish'))
      .pipe(jshint.reporter('fail'));
  };

  es.merge(buildDistJS(), buildTemplates())
    .pipe(plumber({
      errorHandler: handleError
    }))
    .pipe(order([
      'json-formatter.js',
      'json-formatter.html'
    ]))
    .pipe(concat('json-formatter.js'))
    .pipe(header(config.banner, {
      timestamp: (new Date()).toISOString(), pkg: config.pkg
    }))
    .pipe(gulp.dest('dist'))
    .pipe(rename({suffix: '.min'}))
    .pipe(uglify({
        preserveComments: 'some',
        reservedNames: 'Person'
    }))
    .pipe(gulp.dest('./dist'))
    .pipe(connect.reload());
});


gulp.task('styles', function() {

  return gulp.src('src/json-formatter.less')
    .pipe(less())
    .pipe(header(config.banner, {
      timestamp: (new Date()).toISOString(), pkg: config.pkg
    }))
    .pipe(gulp.dest('dist'))
    .pipe(minifyCSS())
    .pipe(rename({suffix: '.min'}))
    .pipe(gulp.dest('dist'))
    .pipe(connect.reload());
});

gulp.task('open', function(){
  return open('http://localhost:8080/demo/demo.html');
});

gulp.task('jshint-test', function(){
  return gulp.src('./test/**/*.js').pipe(jshint());
})

gulp.task('karma', function (done) {
  new KarmaServer({
    configFile: __dirname + '/karma.conf.js',
    singleRun: true
  }, done).start();
});

gulp.task('karma-serve', function(done){
  new KarmaServer({
    configFile: __dirname + '/karma.conf.js'
  }, done).start();
});

function handleError(err) {
  console.log(err.toString());
  this.emit('end');
};

gulp.task('build', function(cb) {
  runSequence('clean', 'scripts', 'styles', cb);
});
gulp.task('serve', ['build', 'connect', 'watch', 'open']);
gulp.task('default', ['build', 'test']);
gulp.task('test', ['build', 'jshint-test', 'karma']);
gulp.task('serve-test', ['build', 'watch', 'jshint-test', 'karma-serve']);
