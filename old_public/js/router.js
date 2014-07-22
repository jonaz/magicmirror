/*global define,require */
define([
		'jquery',
		'underscore',
		'backbone',
		//Define all needed views here to make sure they are included when we build;
		'views/weather',
		'views/calendar',
		'views/clock'
		],
		function($, _, Backbone){
			"use strict";
			var Router = Backbone.Router.extend({
				cache: {
					views: []
				},
				register: function (route, name, path) {
					var self = this;
					this.route(route, name, function () {
						var args = arguments;

                        $(".navbar-collapse").removeClass("in").addClass("collapse");

						require([path], function (Module) {
							var options = null;
							var parameters = route.match(/[:\*]\w+/g);

							// Map the route parameters to options for the View.
							if (parameters) {
								options = {};
								_.each(parameters, function(name, index) {
									options[name.substring(1)] = args[index];
								});
							}

							//create a container div for our new view
							if ( $('.main').find('#view_'+name).length === 0){
								$('.main').append($( '<div class="view" id="view_'+name+'"/>' ));
							}

							//Hide all view containers
							$('.main .view').hide();

							//Show the selected view
							$('.main').find('#view_'+name).show();

							//We cache our views so we only need to instantiate once
							self.cache.views[name] = self.cache.views[name] || new Module({el:$('.main').find('#view_'+name).get(0)});
							self.cache.views[name].options = options;
							self.cache.views[name].beforeDisplay();
							self.cache.views[name].display();
							self.cache.views[name].afterDisplay();
						});
					});
				}

			});


			var router = new Router();
			//router.register('(/)','IndexView','views/index');
			//router.register('login','LoginView','views/login');
			//router.register('admin(/:area)(/:id)','AdminView','views/admin');
			//router.register('lists(/:id)','ListView','views/list');
			//router.register('lists/add/:group','AddListView','views/addList');
			//router.register('lists/edit/:list','AddListView','views/addList');

			//logout is not a view. So we need to add it manually
			router.route('logout', 'logout', function () {
				SessionModel.logout();
			});


			var initialize = function(){
				Backbone.history.start();
			};

			return { initialize: initialize };
		}
);
