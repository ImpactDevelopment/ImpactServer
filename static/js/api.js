// A central place to keep client-side implementations of our api.
// Depends on jQuery, the Promise polyfill, the URL polyfill
(function (global, $) {
    // Set the baseUrl dynamically so that testing on localhost isn't hell
    var baseUrl = (function (scheme, host, version) {
        return scheme + "//api." + host + "/v" + version
    })(global.location.protocol, global.location.host, 1)

    function setToken(token) {
        $.withAuth.setToken(token)
    }

    global.api = {
        // get my user object
        me: function() {
            return new Promise(function (resolve, reject) {
                $.withAuth.get({
                    url: baseUrl + "/user/me",
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(errorThrown)
                    },
                    success: function (data, status) {
                        resolve(data)
                    }
                })
            })
        },
        // Add a way check if logged in
        isLoggedIn: function() {
            return !!window.localStorage.getItem("access_token")
        },
        // Add a way check if logged in
        logout: function() {
            window.localStorage.removeItem("access_token")
        },
        // login with either discord token or username + password
        login: function(email, password) {
            var url = baseUrl + "/login/" + (password ? "password" : "discord")
            var fields = {
                "access_token": password ? undefined : email,
                "email": password ? email : undefined,
                "password": password
            }
            return new Promise(function (resolve, reject) {
                $.post({
                    url: url,
                    data: fields,
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(errorThrown)
                    },
                    success: function (data, status) {
                        setToken(data)
                        resolve("logged in")
                    }
                })
            })
        },
        // register an account. fields should be usable as jQuery's ajax body data
        register: function(fields) {
            return new Promise(function (resolve, reject) {
                $.post({
                    url: baseUrl + "/register/token",
                    data: fields,
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(errorThrown)
                    },
                    success: function (data, status) {
                        setToken(data)
                        resolve("registered")
                    }
                })
            })
        },
        confirmPayment: function(orderID) {
            return new Promise(function (resolve, reject) {
                $.post({
                    url: baseUrl + "/paypal/afterpayment",
                    data: '{"orderID":"' + orderID + '"}',
                    dataType: "json",
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(errorThrown)
                    },
                    success: function (data, status) {
                        resolve(data.token)
                    }
                })
            })
        }
    }
})(window, window.jQuery);

// A small jQuery plugin to make ajax requests with our auth header
// depends on localStorage polyfill
(function($) {

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
        $.withAuth[method] = function(url, data, callback, type) {
            if (typeof url === "object") {
                url.type = method
                return $.withAuth(url)
            }
            // Shift arguments if data argument was omitted
            if ($.isFunction(data)) {
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
    $.withAuth.setToken = function(token) {
        window.localStorage.setItem("access_token", token)
    }
})(jQuery);
