({
	appDir: "../",
	baseUrl: "js",
	dir: "../../dist",
	mainConfigFile: '../js/main.js',
	optimizeCss: 'standard',
	optimize: 'uglify2',
	generateSourceMaps: false,
	preserveLicenseComments: false,
	removeCombined: true,
    paths: {
        requireLib: '../vendor/requirejs/require'
    },
    modules: [{
        name: "main",
        include: ["requireLib", "main"]
    }],
    fileExclusionRegExp: "/build|vendor/"
})
