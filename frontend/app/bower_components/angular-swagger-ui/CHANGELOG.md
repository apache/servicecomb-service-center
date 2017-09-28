### 0.4.4 (2017-06-22)

 * Fix #76: paths with regexp
 * Catch exceptions when generating models (partially fixes #77)
 * Fix module swagger1-to-swagger2-converter: deprecated attr is a string
 * Always display response Content-Type select box

### 0.4.3 (2017-03-30)

 * Fix module swagger1-to-swagger2-converter: losing array items type
 * Optimize model generation
 * Add support for map/dictionary in module swagger1-to-swagger2-converter

### 0.4.2 (2017-03-15)

 * Fix POST/PUT with content type 'x-www-form-urlencoded'
 * Fix model generation: could get some duplicates
 * Fix module swagger-auth: it should not embed angular-ui-bootstrap components (#73)
 * Fix module swagger1-to-swagger2-converter: form parameters were lost
 * Fix module markdown: Allow using markdown in responses description
 * Add support for map/dictonary in model definitions
 * Add possibility to navigate within model complex type using links (#34)

### 0.4.1 (2017-01-18)

 * Fix overriting security definitions (#69)
 * Fix swagger1-to-swagger2-converter when using angular 1.6

### 0.4.0 (2016-12-20)

 * Support angularJS 1.6 (#66)
 * Support Swagger external documentation (#68)

### 0.3.5 (2016-12-15)

 * Fix basePath issue (#46)
 * Fix secondary tags (#61)
 * Create module to allow authorizations by API key or Basic auth for Swagger 2.0 specifications (#38)

### 0.3.4 (2016-11-21)

 * Fix Swagger 1.2 converter
 * Add CSS classes to ease customization
 * Create module to allow rendering markdown (#57)

### 0.3.3 (2016-10-04)

 * Display resources in same order as tags are declared in Swagger specification or alphabetically if no tag defined (#56)
 * Fix `allOf` implementation (#55)
 * CSP compliant templates (#54)

### 0.3.2 (2016-06-30)

 * Fix Swagger model generation (float properties not displayed in samples)
 * i18n add Japanese (#44)
 * Display response headers
 * Fix response references (#50)
 * Fix right-to-left display (#53)

### 0.3.1 (2016-03-21)

 * Make Swagger specification URL absolute when passed to Swagger validator (#43)
 * YAML support
 * Allow rendering a Swagger specification from JSON object or YAML string (#32)
 * Support `allOf` in object definitions (#41)
 * i18n (#37)
 * Handle deprecated operations

### 0.3.0 (2016-02-22)

 * Allow defining a custom template to be used by SwaggerUi (#3)
 * Use ngAnnotate
 * Insert templates in SwaggerUi module
 * Auto add modules at startup to avoid executing `app.run(...)`
 * Fix model generation of objects having no properties
 * Fix Swagger 1.2 references (#35)
 * Fix API explorer when using formData (#36)
 * Fix sample JSON of operations having examples

### 0.2.7 (2015-10-28)

 * Something went wrong with bower, new release to force update

### 0.2.6 (2015-10-28)

 * Lost default responses content-type (#27)
 * Remove useless files when using bower (#25)
 * Remove useless package in dev (#28)
 * Fix dev sample Swagger URL at startup (#29)

### 0.2.5 (2015-10-27)

 * Support path parameters (#26)
 * Remove useless files when using bower (#25)
 * Fix default host and baseURL (#24)
 * Fix "Try it out" request URL if Swagger specification changed
 * Fix Swagger 1.2 host if not defined

### 0.2.4 (2015-10-12)

 * Allow disabling or defining a custom Swagger validator (#19)
 * Fix Swagger 1.2 host and baseURL (#21)
 * Upgrade AngularJS (#23)

### 0.2.3 (2015-08-25)

 * Do not allow adding module multiple times (fix #17)
 * Set permalinks optional as it can cause UI being reloaded when using ngRoute if `reloadOnSearch` not set to false
 * Check parameters at startup to give advices

### 0.2.2 (2015-08-21)

 * Allow forcing parser to JSON
 * Throw error if no parser found
 * Include Swagger validator
 * Create module Swagger 1.2 to Swagger 2.0 converter (beta)
 * Fix duplicated elements in UI

### 0.2.1 (2015-08-13)

 * Fix model generation
 * Fix JSON parser

### 0.2 (2015-08-12)

 * Allow creating and using modules
 * Create module to support Swagger external references
 * Create module to support application/xml API explorer responses to display formatted response body
 * Use the global consumes/produces definition if an operation doesn't specify its own
 * Add support for $ref definition on operation parameters
 * Use bootstrap-less-only to remove dependency to jQuery
 * Display HTML description for operations, parameters and responses. If untrusted sources, user MUST include ngSanitize to his app else he has to add trusted-sources parameter to Swagger-ui directive
 * Fix models defintions circular references

### 0.1.6 (2015-07-07)

 * Fix model generation
 * Fix documentation
 * Display host, base path and version

### 0.1.5 (2015-06-24)

 * Enable permalinks
 * Autoload Swagger specification at startup
 * Transfer to Orange OpenSources

### 0.1.4 (2015-05-03)

 * Fix CSS
 * Fix model generation
 * Fix API explorer displayed URL

### 0.1.3 (2015-04-14)

 * Fix for operations having no tags

### 0.1.2 (2015-04-02)

 * Fix CSS
 * Display API explorer query params
 * Add support for enums
 * Fix model generation
 * Fix display on small screens

### 0.1.1 (2015-03-03)

 * Responsive tables
 * Split JS in multiple files
 * Add CSS classes to ease override
