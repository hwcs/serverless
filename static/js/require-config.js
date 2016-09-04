require.config({
  baseUrl: '../',
  packages: [{
            name: "codemirror",
            location: "dist/codemirror",
            main: "lib/codemirror"
          }],
	paths: {
		'jquery': 'dist/js/jquery-3.1.0.min',
		'bootstrap': 'dist/js/bootstrap.min',
    'bootstrap-table': 'dist/js/bootstrap-table.min',
    'bootstrap-table-local': 'dist/js/bootstrap-table-locale-all.min'
	},
	shim: {
    'jquery': {
      init: function () {
        return jquery.noConflict (true);
      },
      exports: 'jQuery'
    },
    'bootstrap': {
        deps: ['jquery']
    },
    'bootstrap-table': {
        deps: ['bootstrap', 'jquery']
    },
    'bootstrap-table-local': {
        deps: ['bootstrap-table', 'jquery']
    },
	}
});

