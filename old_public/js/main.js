/*global require */
(function () {
    "use strict";
    require.config({
        paths: {
            'jquery': '../vendor/jquery/dist/jquery',
            'underscore': '../vendor/underscore/underscore',
            'hbs': '../vendor/require-handlebars-plugin/hbs',
            'backbone': '../vendor/backbone/backbone',
            'templates': '../templates',
            'jquery.bootstrap': '../vendor/bootstrap/dist/js/bootstrap',
			'toastr': '../vendor/toastr/toastr',
            'd3': '../vendor/d3/d3'
        },
        //shim: {
            //'jquery.bootstrap':{
                //deps: ["jquery"]
            //},
            //'backgrid': {
                //deps: ['jquery', 'underscore', 'backbone'
        //'css!vendor/backgrid/backgrid'
                //],
                //exports: 'Backgrid'
            //}
        //},
        hbs: { // optional
            helpers: true,            // default: true
			i18n: false,              // default: false
			templateExtension: 'hbs', // default: 'hbs'
			partialsUrl: ''           // default: ''
        }
    });

    //require(['views/app'], function(AppView) {
    //new AppView;
    //});



    require(['app' ], function (App) {
        App.initialize();
    });

}());
