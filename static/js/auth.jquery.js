// A small jQuery plugin to make ajax requests with our auth header
(function ($) {

    // Private helper
    function getToken() {
        return window.localStorage.getItem("access_token")
    }

    // Let's be a jQuery plugin
    $.extend({
        withAuth: function(url, options) {
            // Implement jQuery's signatures
            if (typeof url === "object") {
                options = url
                url = undefined
            }
            options = options || {}
            options.url = url || options.url
            options.headers = options.headers || {}

            // Do the auth meme
            if (getToken()) {
                options.headers["Authorization"] = "Bearer " + getToken()
            } else {
                // TODO error out?
            }

            // Hand over to jQuery's method
            return $.ajax(options)
        }
    })

    // add $.withAuth.get() and $.withAuth.post()
    $.each(["get", "post"], function (i, method) {
        $.withAuth[ method ] = function( url, data, callback, type ) {
            // Shift arguments if data argument was omitted
            if ( jQuery.isFunction( data ) ) {
                type = type || callback
                callback = data
                data = undefined
            }

            // Call our plugin
            return $.withAuth({
                url: url,
                type: method,
                dataType: type,
                data: data,
                success: callback
            })
        }
    })

    // Make $.withAuth.ajax() work too
    $.withAuth.ajax = function(url, options) {
        return $.withAuth(url, options)
    }

    // Add a way to set token
    $.withAuth.setToken = function setToken(token) {
        window.localStorage.setItem("access_token", token)
    }
})(jQuery)
