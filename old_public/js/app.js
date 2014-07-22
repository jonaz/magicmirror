/*global define */
define([
		'jquery',
		'backbone',
		'router',
		'toastr',
        //'models/session',
		'jquery.bootstrap'
	],
	function ($, Backbone, Router,Toastr) {
		"use strict";

		return {
			initialize: function(){
                
                //$._parseJSON = $.parseJSON;
                //$.parseJSON = function( data ) {
                    //try {
                        //return $._parseJSON( data );
                    //} catch( err ) {
                        //Toastr.error("Failed JSON decode: "+err.message,'',{positionClass: 'toast-bottom-center'});
                    //}
                //};

                $.ajaxSetup({
                    converters: {
                        "text json": function ( data ) {
                            try {
                                return $.parseJSON( data );
                            } catch( err ) {
                                Toastr.error("Failed JSON decode: "+err.message,'',{positionClass: 'toast-bottom-center'});
                                console.log(err);
                            }
                        }
                    }
                });

                $(document).ajaxError(function( event, xhr, settings ) {
                    try {
                        var ret = JSON.parse(xhr.responseText);
                        var parser = document.createElement('a');
                        parser.href = settings.url;

                        if ( ret.error !== undefined ) {
                            Toastr.error("<strong>"+parser.pathname+"</strong><br />"+xhr.status + ": " + ret.error,'',{positionClass: 'toast-bottom-center'});
                            console.log( xhr.status + ": " + ret.error + " ("+settings.url+")" );
                        } else {
                            Toastr.error(xhr.status + ": ",'',{positionClass: 'toast-bottom-center'});
                        }
                    } catch ( err ) {
                        console.log(err);
                    }

                    //if ( xhr.status === 401 ) {
                        //if ( Backbone.history.getHash() !== "login" ) {
                            //sessionModel.logout(function() {
                                //Backbone.history.navigate('login', { trigger : true });
                            //});
                        //}
                    //}
                });

				Router.initialize();
			}
		};

});
