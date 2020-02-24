/* http://www.minifier.org */
(function($){function getToken(){return window.localStorage.getItem("access_token")}
    $.extend({withAuth:function(url,options){if(typeof url==="object"){options=url
            url=undefined}
            options=options||{}
            options.url=url||options.url
            options.headers=options.headers||{}
            if(getToken()){options.headers.Authorization="Bearer "+getToken()}else{}
            return $.ajax(options)}})
    $.each(["get","post"],function(i,method){$.withAuth[method]=function(url,data,callback,type){if(jQuery.isFunction(data)){type=type||callback
        callback=data
        data=undefined}
        return $.withAuth({url:url,type:method,dataType:type,data:data,success:callback})}})
    $.withAuth.ajax=function(url,options){return $.withAuth(url,options)}
    $.withAuth.setToken=function setToken(token){window.localStorage.setItem("access_token",token)}})(jQuery)