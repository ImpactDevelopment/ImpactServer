// A central place to keep client-side implementations of our api.
// Depends on jQuery, the Promise polyfill, the URL polyfill
(function (global, $) {
    // Set the baseUrl dynamically so that testing on localhost isn't hell
    var baseUrl = (function (scheme, host, version) {
        return scheme + "//api." + host + "/v" + version
    })(global.location.protocol, global.location.host, 1)

    function messageFromjqXHR(jqXHR) {
        try {
            return JSON.parse(jqXHR.responseText).message
        } catch (e) {
            return jqXHR.responseText
        }
    }

    function addDashesToUUID(id) {
        // Sanitize first, then add dashes where we want them
        id = id.replace(/-/g, "")
        return id.substr(0,8)+"-"+id.substr(8,4)+"-"+id.substr(12,4)+"-"+id.substr(16,4)+"-"+id.substr(20)
    }

    global.api = {
        // if data is provided, patches the user otherwise gets the user info
        me: function(data) {
            return new Promise(function (resolve, reject) {
                $.withAuth({
                    url: baseUrl + "/user/me",
                    method: data ? "PATCH" : "GET",
                    // echo sometimes fails to bind form data to *bool, so send json
                    data: data ? JSON.stringify(data) : undefined,
                    headers: data ? {"Content-Type": "application/json"} : undefined,
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(messageFromjqXHR(jqXHR))
                    },
                    success: function (result, status) {
                        api.user = result
                        resolve(result)
                    }
                })
            })
        },
        // Add a way check if logged in
        isLoggedIn: function() {
            return !!window.localStorage.getItem("access_token")
        },
        // True if the account has email set
        isFullAccount: function(user) {
            if (!user) user = api.user
            if (!user) return false
            return !!user.email
        },
        // Add a way check if logged in
        logout: function() {
            window.localStorage.removeItem("access_token")
        },
        setToken: function(token) {
            $.withAuth.setToken(token)
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
                        reject(messageFromjqXHR(jqXHR))
                    },
                    success: function (data, status) {
                        api.setToken(data)
                        resolve("logged in")
                    }
                })
            })
        },
        // register an account. fields should be usable as jQuery's ajax body data
        register: function(fields) {
            var post = api.isLoggedIn() ? $.withAuth.post : $.post

            return new Promise(function (resolve, reject) {
                post({
                    url: baseUrl + "/register/token",
                    data: fields,
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(errorThrown + ": " + messageFromjqXHR(jqXHR))
                    },
                    success: function (data, status) {
                        api.setToken(data)
                        resolve("registered")
                    }
                })
            })
        },
        getTokenInfo: function(token) {
            return new Promise(function (resolve, reject) {
                $.get({
                    url: baseUrl + "/checktoken?token=" + encodeURIComponent(token),
                    dataType: "json",
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(errorThrown + ": " + messageFromjqXHR(jqXHR))
                    },
                    success: function (data, status) {
                        resolve(data)
                    }
                })
            })
        },
        stripeInfo: function() {
            return new Promise(function(resolve, reject) {
                $.get({
                    url: baseUrl + "/stripe/info",
                    dataType: "json",
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(messageFromjqXHR(jqXHR))
                    },
                    success: function (data, status) {
                        resolve(data)
                    }
                })
            })
        },
        createPayment: function(currency, amount, email, verification) {
            if (!email) {
                // currency is optional, if only two args are present then shift them right
                email = amount
                amount = currency
                currency = undefined
            }
            return new Promise(function(resolve, reject) {
                $.post({
                    url: baseUrl + "/stripe/createpayment",
                    data: {
                        currency: currency,
                        amount: amount,
                        email: email,
                        "g-recaptcha-response": verification
                    },
                    dataType: "json",
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(messageFromjqXHR(jqXHR))
                    },
                    success: function (data, status) {
                        resolve(data)
                    }
                })
            })
        },
        redeemPayment: function(paymentID, email) {
            return new Promise(function (resolve, reject) {
                $.post({
                    url: baseUrl + "/stripe/redeem",
                    data: {
                        'payment_id': paymentID,
                        email: email
                    },
                    dataType: "json",
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(messageFromjqXHR(jqXHR))
                    },
                    success: function (data, status) {
                        resolve(data.token)
                    }
                })
            })
        },
        // returns a one-time login link url for Stripe Connect's Express Account dashboard
        stripeConnectLogin: function() {
            return new Promise(function(resolve, reject) {
                $.withAuth.get({
                    url: baseUrl + "/stripe/connect/login",
                    dataType: "json",
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(messageFromjqXHR(jqXHR))
                    },
                    success: function (data, status) {
                        resolve(data)
                    }
                })
            })
        },
        forgotPassword: function(email, verification) {
            return new Promise(function (resolve, reject) {
                $.post({
                    url: baseUrl + "/password/reset",
                    data: {
                        email: email,
                        "g-recaptcha-response": verification
                    },
                    dataType: "json",
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(messageFromjqXHR(jqXHR))
                    },
                    success: function (data, status) {
                        resolve(data.message ? data.message : data)
                    }
                })
            })
        },
        // Change password using a reset token. If logged in then token can optionally be omitted.
        changePassword: function(token, password) {
            if (!password) {
                password = token
                token = undefined
            }
            var url = token ? baseUrl + "/password/" + encodeURIComponent(token) : baseUrl + "/password/me"
            var ajax = token ? $.ajax : $.withAuth

            return new Promise(function (resolve, reject) {
                ajax({
                    url: url,
                    method: "PUT",
                    data: {
                        token: token,
                        password: password
                    },
                    dataType: "json",
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(messageFromjqXHR(jqXHR))
                    },
                    success: function (data, status) {
                        resolve(data.message ? data.message : data)
                    }
                })
            })
        },
        /*
            {
                id: the user's id
                email: the user's account email
                username: the user's username (not nickname)
                tag: the user's #1234 tag
                avatar: the user's avatar url (PNG or GIF)
                nitro: whether or not the user has nitro
            }
         */
        getDiscordUser: function (token) {
            // TODO get from user info if token is omitted
            return new Promise(function (resolve, reject) {

                $.get({
                    url: "https://discordapp.com/api/v6/users/@me",
                    dataType: "json",
                    headers: {
                        // https://discordapp.com/developers/docs/reference#http-api
                        "Authorization": "Bearer " + token
                    },
                    error: function (jqXHR, textStatus, errorThrown) {
                        reject(errorThrown)
                    },
                    success: function (data, status) {
                        // get avatar url https://discordapp.com/developers/docs/reference#image-formatting
                        var avatar = ""
                        if (data && data.avatar && data.id) {
                            var ext = "a_" === data.avatar.substring(0, 2) ? ".gif" : ".png"
                            var base = "https://cdn.discordapp.com/avatars/"
                            avatar = base + data.id + "/" + data.avatar + ext
                        } else if (data && data.discriminator) {
                            // Users without avatars use a "Default User Avatar"
                            avatar = "https://cdn.discordapp.com/embed/avatars/" + (data.discriminator % 5) + ".png"
                        }

                        // Pass a serialize the user object
                        resolve({
                            id: data.id || "",
                            email: data.email || "",
                            username: data.username || "",
                            discriminator: data.discriminator || "",
                            avatar: avatar,
                            nitro: data.premium_type && data.premium_type > 0
                        })
                    }
                })
            })
        },
        getMinecraftUser: function(name) {
            // Get the user profile from mcheads or minetools
            // luckily both APIs have name and id json properties and accept either name or uuid input
            var main = "https://api.minetools.eu/uuid/"
            var fallback = "https://mc-heads.net/minecraft/profile/"
            return new Promise(function (resolve, reject) {
                // Generate a function here to aid recursion when fallback is enabled
                // base is either main or fallback
                function req(base) {
                    $.get({
                        url: base + encodeURIComponent(("" + name).trim()),
                        dataType: "json",
                        error: function (jqXHR, textStatus, errorThrown) {
                            // If fallback is enabled, recurse one level
                            if (base !== fallback) {
                                console.warn("MineTools API failed, falling back to MC Heads for UUID lookup")
                                req(fallback)
                            } else {
                                reject(errorThrown)
                            }
                        },
                        success: function (data, status) {
                            if (!data ||
                                !data.id || data.id === "null" ||
                                (data.status && (""+data.status).toLowerCase() !== "ok")) {
                                reject("No user found")
                            } else {
                                resolve({
                                    id: addDashesToUUID(data.id),
                                    name: data.name
                                })
                            }
                        }
                    })
                }

                req(main)
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
