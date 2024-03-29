var c_ajax_function = null;

function c_set_window_title(title) {
  if (!title) {
    title = "Portal";
  }
  document.title = title + " - Qualitec";
}

function show_loading_info() {
  center_v("#c-loading-info");
  $("#c-loading-info").show();
  $("#c-loading-info").css("z-index", 50);
}

function hide_loading_info() {
  $("#c-loading-info").hide();
}

window.onpopstate = function (event) {
  c_ajax_do("get", event.state.uri, null, "replace");
};

function adjust_loaded_page(obj) {
  obj.find(".c-ajax,[data-celus='ajax']").each(function () {
    switch ($(this).prop("tagName")) {
      case "SELECT":
        $(this).on("change", function (event) {
          if (!get_tab($(this).val(), event)) {
            event.preventDefault();
          }
        });
        break;

      case "FORM":
        $(this).submit(function (event) {
          // Limpa mensagens de erro do formulário
          $(this).find(".invalid-feedback").remove();
          $(this).find(".is-invalid").removeClass("is-invalid");

          const func = __find_function($(this).attr("func-c-dest"));

          c_ajax_do(
            $(this).attr("method"),
            $(this).attr("action"),
            $(this).serialize(),
            $(this).attr("data-celus-history"),
            func
          );

          event.preventDefault();
        });
        break;

      case "BUTTON":
        target = $(this).attr("target");
        if (typeof target !== "undefined" && target !== false) {
          $(this).on("click", function (event) {
            $("input[name='" + target + "']").val($(this).val());
            console.log($(this).val());
          });
        } else {
          $(this).on("click", function (event) {
            c_ajax_load($(this));
            event.preventDefault();
          });
        }
        break;

      default:
        $(this).on("click", function (event) {
          if (!get_tab($(this).attr("href"), event)) {
            event.preventDefault();
          }
        });
    }
  });

  obj.find(".c-load, [data-c-load='polling']").each(function () {
    c_ajax_load($(this));
  });

  // Novo modelo de buscas
  obj.find("form.c-search").each(function () {
    $(this).submit(function (event) {
      var input_control = $(this).find("input[name='q']");
      if (input_control.length) {
        var u = new URI(window.location);
        if (input_control.val()) {
          u.setSearch("q", input_control.val());
        } else {
          u.removeSearch("q");
        }
        get_tab(u.toString());
      }
      event.preventDefault();
    });
  });

  c_set_window_title($("#c-title").text());

  // Ajusta os links do menu lateral
  var allSites = getSelectedSites(new URI(window.location));
  updateMenuLateral(allSites);
  updateSiteBar(allSites);

  // Ajusta a página ativa no menu
  $("#mainnav ul li.active").each(function () {
    $(this).removeClass("active");
  });
  $("#mainnav ul li [href]").each(function () {
    if (
      $(this).attr("href").split("/")[1].split("?")[0] ===
      $(location).prop("pathname").split("/")[1]
    ) {
      // Exceção para tratar o Novo Chamado
      if ($(this).attr("href").split("?")[0] === "/tickets/new") {
        if ($(location).prop("pathname") === "/tickets/new") {
          $(this).parent().addClass("active");
          return false;
        }
      } else {
        $(this).parent().addClass("active");
        return false;
      }
    }
  });
}

function center_v(selector) {
  $(selector).css({
    left: "50%",
    "margin-left": function () {
      return -$(this).outerWidth() / 2;
    },
  });
}

function center_vh(obj) {
  obj.css({
    left: "50%",
    top: "50%",
    "margin-left": function () {
      return -$(this).outerWidth() / 2;
    },
    "margin-top": function () {
      return -$(this).outerHeight() / 2;
    },
  });
}

function show_success_message(message) {
  $("body").append(
    '<div class="alert alert-success alert-dismissible" role="alert" id="success-alert">' +
      '<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button>' +
      "<strong>" +
      message +
      "</strong>" +
      "</div>"
  );

  center_vh($("#success-alert"));
  $("#success-alert")
    .fadeTo("slow", 1)
    .delay(3000)
    .fadeTo("slow", 0, function () {
      $("#success-alert").alert("close");
    });
}

function get_tab(uri, e) {
  if (c_ajax_function && !c_ajax_function()) {
    return false;
  }
  if (e && e.ctrlKey) {
    window.open(uri, "_blank");
  } else {
    c_ajax_do("get", uri, null, null);
  }
  return false;
}

var in_ajax = false;

function c_ajax_set_page(uri, data, history_mode) {
  obj = $("#content");
  obj.html(data);
  if (history_mode === "replace") {
    history.replaceState({ uri: uri }, null, uri);
  } else {
    history.pushState({ uri: uri }, null, uri);
  }
  adjust_loaded_page(obj);
}

function c_ajax_process_response_array(
  uri,
  response_array,
  history_mode,
  func
) {
  if (!Array.isArray(response_array)) {
    if (func) {
      func(response_array);
    }
    return;
  }
  for (i = 0; i < response_array.length; ++i) {
    switch (response_array[i][0]) {
      case 0: // html data
        if (func) {
          func(response_array);
        } else {
          c_ajax_set_page(uri, response_array[i][1], history_mode);
        }
        break;
      case 10: // Mensagem de sucesso
        show_success_message(response_array[i][1]);
        break;
      case 30: // Redirecionamento GET
        in_ajax = false;
        c_ajax_do("GET", response_array[i][1], null, history_mode, func);
        break;
    }
  }
}

function c_ajax_process_response_error(jqXHR) {
  if (jqXHR.getResponseHeader("x-celus-error") === "1") {
    var error = JSON.parse(jqXHR.responseText);
    var general_errors = [];
    for (var i = 0; i < error.length; i++) {
      switch (error[i][0]) {
        case 10:
          general_errors.push(error[i][1]);
          break;

        case 20:
          var control = $("#" + error[i][1]);
          if (control.length) {
            control.addClass("is-invalid");
            if (control.parent().children(".invalid-feedback").length) {
              control.parent().children(".invalid-feedback").text(error[i][2]);
            } else {
              control
                .parent()
                .append(
                  "<span class='invalid-feedback'>" + error[i][2] + "</span>"
                );
            }
          } else {
            general_errors.push(error[i][1] + ": " + error[i][2]);
          }
          break;
      }
    }

    if (general_errors.length > 0) {
      var error_html = "";
      $.each(general_errors, function (i, val) {
        error_html += "<li>" + val + "</li>";
      });
      $("#error-info .modal-body").html("<ul>" + error_html + "</ul>");
      $("#error-info").modal();
    }
  } else {
    $("#error-info .modal-body").html(jqXHR.responseText);
    $("#error-info").modal();
  }
}

function __find_function(func_name) {
  if (func_name) {
    for (let i in window) {
      if (typeof window[i] === "function" && window[i].name === func_name) {
        return window[i];
      }
    }
  }
  return undefined;
}

function c_ajax_do(type, uri, data, history_mode, func) {
  if (in_ajax) {
    return false;
  }
  in_ajax = true;

  show_loading_info();

  console.debug(type);
  $.ajax({
    type: type || "post",
    url: uri,
    cache: false,
    data: data,
  })
    .done(function (data, textStatus, jqXHR) {
      console.debug(data);
      if (jqXHR.responseJSON) {
        c_ajax_process_response_array(
          uri,
          jqXHR.responseJSON,
          history_mode,
          func
        );
      } else {
        c_ajax_set_page(uri, data, history_mode);
      }
    })

    .fail(function (jqXHR) {
      if (jqXHR.status === 404) {
        c_ajax_set_page(uri, jqXHR.responseText, null);
      } else {
        c_ajax_process_response_error(jqXHR);
      }
    })

    .always(function () {
      hide_loading_info();
      in_ajax = false;
    });
}

function c_ajax_load(obj) {
  let dest = obj.attr("data-c-dest");

  if (dest) {
    dest = $("#" + dest);
  } else {
    dest = __find_function(obj, func_namobj.attr("func-c-dest"));
  }

  if (!dest) {
    dest = obj;
  }

  $.ajax({
    type: "GET",
    url: obj.attr("data-c-url"),
    cache: false,
    beforeSend: function (jqXHR, settings) {
      stamp = obj.attr("data-c-stamp");
      if (stamp) {
        jqXHR.setRequestHeader("x-celus-stamp", stamp);
      }
      // Apenas quando não foir polling exibe a ampulheta de carregando
      if (
        dest &&
        typeof dest !== "function" &&
        obj.attr("data-c-load") !== "polling"
      ) {
        dest.prepend(
          '<span style="text-align: center; position:absolute; width: 100px; left: 50%; margin-left: -50px;"><i class="fa fa-refresh fa-spin fa-2x fa-fw"></i>'
        );
      }
    },
  })
    .done(function (data, textStatus, jqXHR) {
      if (jqXHR.getResponseHeader("x-celus-timeout") !== "1") {
        if (typeof dest === "function") {
          console.log("Is a function");
          dest(data);
          return;
        } else {
          dest.html(data);
        }
      }
      stamp = jqXHR.getResponseHeader("x-celus-stamp");
      if (stamp) {
        dest.attr("data-c-stamp", stamp);
      }

      adjust_loaded_page(dest);

      if (obj.attr("data-c-load") === "polling") {
        c_ajax_load(obj);
      }
    })
    .fail(function (jqXHR) {
      dest.text(jqXHR.responseText);
    });
}

var drop = null;
var dropDown = null;
var dropDownMenu = null;
var input = null;
var list = null;
var sitesSelecteds = null;

function initBase() {
  drop = $("#site-menu");
  dropDown = drop.find(".dropdown");
  dropDownMenu = $("#site-modal");
  input = dropDownMenu.find("input");
  list = dropDownMenu.find("table");
  sitesSelecteds = drop.find(".sites-selecteds");

  dropDownMenu.on("shown.bs.modal", function () {
    input.focus();
    input.val("");
    input.trigger("input");
    list.find("tr:visible:first").addClass("table-active");
  });

  list.find("tr").each(function () {
    $(this).on("click", function (event) {
      event.preventDefault();
      addSite($(this).data("id"), $(this).find("td").text());
    });
    $(this).css("cursor", "pointer");
  });

  input.keydown(function (event) {
    // Seta para baixo e seta para cima. Move a seleção.
    // TODO: mover scroll
    if (event.which === 40 || event.which === 38) {
      event.preventDefault();
      list.find("tr.table-active:first").each(function () {
        var e =
          event.which === 40
            ? $(this).nextAll(":visible:first")
            : $(this).prevAll(":visible:first");
        if (e.length) {
          $(this).removeClass("table-active");
          e.addClass("table-active");
        }
      });

      // Enter confirma a seleção e adiciona na lista de sites
    } else if (event.which == 13) {
      event.preventDefault();
      var sel = list.find("tr.table-active:visible:first");
      if (sel.length > 0) {
        addSite(sel.data("id"), sel.find("td").text());
      }
    }
  });

  input.on("input", function () {
    var text = $(this).val().toLowerCase();
    var i = 0;
    var selectedSites = getSelectedSites();
    list.find("tr").each(function () {
      var td = $(this).find("td");
      $(this).removeClass("table-active");
      td.unmark();
      if (
        td.text().toLowerCase().indexOf(text) >= 0 &&
        selectedSites.indexOf($(this).data("id").toString()) == -1
      ) {
        td.mark(text);
        $(this).show();
        if (i === 0) {
          $(this).addClass("table-active");
        }
        i++;
      } else {
        $(this).hide();
      }
    });
  });
}

function addSite(id, text) {
  var u = new URI(window.location);
  console.log(u);
  u.addSearch("s", id);
  get_tab(u.toString());
  dropDownMenu.modal("hide"); //.removeClass('show');
}

function removeSite(id) {
  var u = new URI(window.location);
  u.removeSearch("s", id);
  get_tab(u.toString());
}

function updateMenuLateral(allSites) {
  $(".dual-nav").collapse("hide");
  $("#mainnav ul li [href]").each(function () {
    var u = new URI($(this).attr("href"));
    u.removeSearch("s");
    for (var i = 0; i < allSites.length; i++) {
      u.addSearch("s", allSites[i]);
    }
    $(this).attr("href", u.toString());
  });
}

function updateSiteBar(allSites) {
  var filter = '[data-id="' + allSites.join('"],[data-id="') + '"]';
  sitesSelecteds.find("span").not(filter).remove();
  var curSites = sitesSelecteds
    .find("span")
    .map(function () {
      return $(this).data("id").toString();
    })
    .get();
  var addSites = allSites.filter(function (item) {
    return curSites.indexOf(item) === -1;
  });
  for (var i = 0; i < addSites.length; ++i) {
    var e = list.find('[data-id="' + addSites[i] + '"] td');
    if (e.length > 0) {
      sitesSelecteds.append(
        '<span class="badge badge-secondary mr-1" data-id="' +
          addSites[i] +
          '">' +
          list.find('[data-id="' + addSites[i] + '"] td').text() +
          ' <a href="#" style="color: white" onclick="removeSite(' +
          addSites[i] +
          ');return false;"><i class="fa fa-times" aria-hidden="true"></i></a></span>'
      );
    }
  }
}

function getSelectedSites(u) {
  if (typeof u === "undefined") {
    u = new URI(window.location);
  }
  var sel = u.hasQuery("s") ? u.search(true)["s"] : [];
  // Quando há apenas 1 query 's' o objeto URI retorna uma
  // string com a query ao invés de um array com um único elento.
  // Assim, para manter o código coeso, ela é convertida em array.
  if (typeof sel === "string") {
    sel = [sel];
  }
  return sel.filter(String);
}

/*! URI.js v1.19.1 http://medialize.github.io/URI.js/ */
/* build contains: URI.js */
/*
 URI.js - Mutating URLs

 Version: 1.19.1

 Author: Rodney Rehm
 Web: http://medialize.github.io/URI.js/

 Licensed under
   MIT License http://www.opensource.org/licenses/mit-license

*/
(function (m, v) {
  "object" === typeof module && module.exports
    ? (module.exports = v(
        require("./punycode"),
        require("./IPv6"),
        require("./SecondLevelDomains")
      ))
    : "function" === typeof define && define.amd
    ? define(["./punycode", "./IPv6", "./SecondLevelDomains"], v)
    : (m.URI = v(m.punycode, m.IPv6, m.SecondLevelDomains, m));
})(this, function (m, v, t, h) {
  function d(a, b) {
    var c = 1 <= arguments.length,
      e = 2 <= arguments.length;
    if (!(this instanceof d)) return c ? (e ? new d(a, b) : new d(a)) : new d();
    if (void 0 === a) {
      if (c) throw new TypeError("undefined is not a valid argument for URI");
      a = "undefined" !== typeof location ? location.href + "" : "";
    }
    if (null === a && c)
      throw new TypeError("null is not a valid argument for URI");
    this.href(a);
    return void 0 !== b ? this.absoluteTo(b) : this;
  }
  function p(a) {
    return a.replace(/([.*+?^=!:${}()|[\]\/\\])/g, "\\$1");
  }
  function u(a) {
    return void 0 === a
      ? "Undefined"
      : String(Object.prototype.toString.call(a)).slice(8, -1);
  }
  function k(a) {
    return "Array" === u(a);
  }
  function C(a, b) {
    var c = {},
      d;
    if ("RegExp" === u(b)) c = null;
    else if (k(b)) {
      var f = 0;
      for (d = b.length; f < d; f++) c[b[f]] = !0;
    } else c[b] = !0;
    f = 0;
    for (d = a.length; f < d; f++)
      if ((c && void 0 !== c[a[f]]) || (!c && b.test(a[f])))
        a.splice(f, 1), d--, f--;
    return a;
  }
  function w(a, b) {
    var c;
    if (k(b)) {
      var d = 0;
      for (c = b.length; d < c; d++) if (!w(a, b[d])) return !1;
      return !0;
    }
    var f = u(b);
    d = 0;
    for (c = a.length; d < c; d++)
      if ("RegExp" === f) {
        if ("string" === typeof a[d] && a[d].match(b)) return !0;
      } else if (a[d] === b) return !0;
    return !1;
  }
  function D(a, b) {
    if (!k(a) || !k(b) || a.length !== b.length) return !1;
    a.sort();
    b.sort();
    for (var c = 0, d = a.length; c < d; c++) if (a[c] !== b[c]) return !1;
    return !0;
  }
  function z(a) {
    return a.replace(/^\/+|\/+$/g, "");
  }
  function F(a) {
    return escape(a);
  }
  function A(a) {
    return encodeURIComponent(a)
      .replace(/[!'()*]/g, F)
      .replace(/\*/g, "%2A");
  }
  function x(a) {
    return function (b, c) {
      if (void 0 === b) return this._parts[a] || "";
      this._parts[a] = b || null;
      this.build(!c);
      return this;
    };
  }
  function E(a, b) {
    return function (c, d) {
      if (void 0 === c) return this._parts[a] || "";
      null !== c && ((c += ""), c.charAt(0) === b && (c = c.substring(1)));
      this._parts[a] = c;
      this.build(!d);
      return this;
    };
  }
  var G = h && h.URI;
  d.version = "1.19.1";
  var g = d.prototype,
    l = Object.prototype.hasOwnProperty;
  d._parts = function () {
    return {
      protocol: null,
      username: null,
      password: null,
      hostname: null,
      urn: null,
      port: null,
      path: null,
      query: null,
      fragment: null,
      preventInvalidHostname: d.preventInvalidHostname,
      duplicateQueryParameters: d.duplicateQueryParameters,
      escapeQuerySpace: d.escapeQuerySpace,
    };
  };
  d.preventInvalidHostname = !1;
  d.duplicateQueryParameters = !1;
  d.escapeQuerySpace = !0;
  d.protocol_expression = /^[a-z][a-z0-9.+-]*$/i;
  d.idn_expression = /[^a-z0-9\._-]/i;
  d.punycode_expression = /(xn--)/i;
  d.ip4_expression = /^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/;
  d.ip6_expression =
    /^\s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?\s*$/;
  d.find_uri_expression =
    /\b((?:[a-z][\w-]+:(?:\/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}\/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s`!()\[\]{};:'".,<>?\u00ab\u00bb\u201c\u201d\u2018\u2019]))/gi;
  d.findUri = {
    start: /\b(?:([a-z][a-z0-9.+-]*:\/\/)|www\.)/gi,
    end: /[\s\r\n]|$/,
    trim: /[`!()\[\]{};:'".,<>?\u00ab\u00bb\u201c\u201d\u201e\u2018\u2019]+$/,
    parens: /(\([^\)]*\)|\[[^\]]*\]|\{[^}]*\}|<[^>]*>)/g,
  };
  d.defaultPorts = {
    http: "80",
    https: "443",
    ftp: "21",
    gopher: "70",
    ws: "80",
    wss: "443",
  };
  d.hostProtocols = ["http", "https"];
  d.invalid_hostname_characters = /[^a-zA-Z0-9\.\-:_]/;
  d.domAttributes = {
    a: "href",
    blockquote: "cite",
    link: "href",
    base: "href",
    script: "src",
    form: "action",
    img: "src",
    area: "href",
    iframe: "src",
    embed: "src",
    source: "src",
    track: "src",
    input: "src",
    audio: "src",
    video: "src",
  };
  d.getDomAttribute = function (a) {
    if (a && a.nodeName) {
      var b = a.nodeName.toLowerCase();
      if ("input" !== b || "image" === a.type) return d.domAttributes[b];
    }
  };
  d.encode = A;
  d.decode = decodeURIComponent;
  d.iso8859 = function () {
    d.encode = escape;
    d.decode = unescape;
  };
  d.unicode = function () {
    d.encode = A;
    d.decode = decodeURIComponent;
  };
  d.characters = {
    pathname: {
      encode: {
        expression: /%(24|26|2B|2C|3B|3D|3A|40)/gi,
        map: {
          "%24": "$",
          "%26": "&",
          "%2B": "+",
          "%2C": ",",
          "%3B": ";",
          "%3D": "=",
          "%3A": ":",
          "%40": "@",
        },
      },
      decode: {
        expression: /[\/\?#]/g,
        map: { "/": "%2F", "?": "%3F", "#": "%23" },
      },
    },
    reserved: {
      encode: {
        expression:
          /%(21|23|24|26|27|28|29|2A|2B|2C|2F|3A|3B|3D|3F|40|5B|5D)/gi,
        map: {
          "%3A": ":",
          "%2F": "/",
          "%3F": "?",
          "%23": "#",
          "%5B": "[",
          "%5D": "]",
          "%40": "@",
          "%21": "!",
          "%24": "$",
          "%26": "&",
          "%27": "'",
          "%28": "(",
          "%29": ")",
          "%2A": "*",
          "%2B": "+",
          "%2C": ",",
          "%3B": ";",
          "%3D": "=",
        },
      },
    },
    urnpath: {
      encode: {
        expression: /%(21|24|27|28|29|2A|2B|2C|3B|3D|40)/gi,
        map: {
          "%21": "!",
          "%24": "$",
          "%27": "'",
          "%28": "(",
          "%29": ")",
          "%2A": "*",
          "%2B": "+",
          "%2C": ",",
          "%3B": ";",
          "%3D": "=",
          "%40": "@",
        },
      },
      decode: {
        expression: /[\/\?#:]/g,
        map: { "/": "%2F", "?": "%3F", "#": "%23", ":": "%3A" },
      },
    },
  };
  d.encodeQuery = function (a, b) {
    var c = d.encode(a + "");
    void 0 === b && (b = d.escapeQuerySpace);
    return b ? c.replace(/%20/g, "+") : c;
  };
  d.decodeQuery = function (a, b) {
    a += "";
    void 0 === b && (b = d.escapeQuerySpace);
    try {
      return d.decode(b ? a.replace(/\+/g, "%20") : a);
    } catch (c) {
      return a;
    }
  };
  var r = { encode: "encode", decode: "decode" },
    y,
    B = function (a, b) {
      return function (c) {
        try {
          return d[b](c + "").replace(
            d.characters[a][b].expression,
            function (c) {
              return d.characters[a][b].map[c];
            }
          );
        } catch (e) {
          return c;
        }
      };
    };
  for (y in r)
    (d[y + "PathSegment"] = B("pathname", r[y])),
      (d[y + "UrnPathSegment"] = B("urnpath", r[y]));
  r = function (a, b, c) {
    return function (e) {
      var f = c
        ? function (a) {
            return d[b](d[c](a));
          }
        : d[b];
      e = (e + "").split(a);
      for (var g = 0, n = e.length; g < n; g++) e[g] = f(e[g]);
      return e.join(a);
    };
  };
  d.decodePath = r("/", "decodePathSegment");
  d.decodeUrnPath = r(":", "decodeUrnPathSegment");
  d.recodePath = r("/", "encodePathSegment", "decode");
  d.recodeUrnPath = r(":", "encodeUrnPathSegment", "decode");
  d.encodeReserved = B("reserved", "encode");
  d.parse = function (a, b) {
    b || (b = { preventInvalidHostname: d.preventInvalidHostname });
    var c = a.indexOf("#");
    -1 < c &&
      ((b.fragment = a.substring(c + 1) || null), (a = a.substring(0, c)));
    c = a.indexOf("?");
    -1 < c && ((b.query = a.substring(c + 1) || null), (a = a.substring(0, c)));
    "//" === a.substring(0, 2)
      ? ((b.protocol = null),
        (a = a.substring(2)),
        (a = d.parseAuthority(a, b)))
      : ((c = a.indexOf(":")),
        -1 < c &&
          ((b.protocol = a.substring(0, c) || null),
          b.protocol && !b.protocol.match(d.protocol_expression)
            ? (b.protocol = void 0)
            : "//" === a.substring(c + 1, c + 3)
            ? ((a = a.substring(c + 3)), (a = d.parseAuthority(a, b)))
            : ((a = a.substring(c + 1)), (b.urn = !0))));
    b.path = a;
    return b;
  };
  d.parseHost = function (a, b) {
    a || (a = "");
    a = a.replace(/\\/g, "/");
    var c = a.indexOf("/");
    -1 === c && (c = a.length);
    if ("[" === a.charAt(0)) {
      var e = a.indexOf("]");
      b.hostname = a.substring(1, e) || null;
      b.port = a.substring(e + 2, c) || null;
      "/" === b.port && (b.port = null);
    } else {
      var f = a.indexOf(":");
      e = a.indexOf("/");
      f = a.indexOf(":", f + 1);
      -1 !== f && (-1 === e || f < e)
        ? ((b.hostname = a.substring(0, c) || null), (b.port = null))
        : ((e = a.substring(0, c).split(":")),
          (b.hostname = e[0] || null),
          (b.port = e[1] || null));
    }
    b.hostname && "/" !== a.substring(c).charAt(0) && (c++, (a = "/" + a));
    b.preventInvalidHostname && d.ensureValidHostname(b.hostname, b.protocol);
    b.port && d.ensureValidPort(b.port);
    return a.substring(c) || "/";
  };
  d.parseAuthority = function (a, b) {
    a = d.parseUserinfo(a, b);
    return d.parseHost(a, b);
  };
  d.parseUserinfo = function (a, b) {
    var c = a.indexOf("/"),
      e = a.lastIndexOf("@", -1 < c ? c : a.length - 1);
    -1 < e && (-1 === c || e < c)
      ? ((c = a.substring(0, e).split(":")),
        (b.username = c[0] ? d.decode(c[0]) : null),
        c.shift(),
        (b.password = c[0] ? d.decode(c.join(":")) : null),
        (a = a.substring(e + 1)))
      : ((b.username = null), (b.password = null));
    return a;
  };
  d.parseQuery = function (a, b) {
    if (!a) return {};
    a = a.replace(/&+/g, "&").replace(/^\?*&*|&+$/g, "");
    if (!a) return {};
    for (var c = {}, e = a.split("&"), f = e.length, g, n, k = 0; k < f; k++)
      if (
        ((g = e[k].split("=")),
        (n = d.decodeQuery(g.shift(), b)),
        (g = g.length ? d.decodeQuery(g.join("="), b) : null),
        l.call(c, n))
      ) {
        if ("string" === typeof c[n] || null === c[n]) c[n] = [c[n]];
        c[n].push(g);
      } else c[n] = g;
    return c;
  };
  d.build = function (a) {
    var b = "";
    a.protocol && (b += a.protocol + ":");
    a.urn || (!b && !a.hostname) || (b += "//");
    b += d.buildAuthority(a) || "";
    "string" === typeof a.path &&
      ("/" !== a.path.charAt(0) && "string" === typeof a.hostname && (b += "/"),
      (b += a.path));
    "string" === typeof a.query && a.query && (b += "?" + a.query);
    "string" === typeof a.fragment && a.fragment && (b += "#" + a.fragment);
    return b;
  };
  d.buildHost = function (a) {
    var b = "";
    if (a.hostname)
      b = d.ip6_expression.test(a.hostname)
        ? b + ("[" + a.hostname + "]")
        : b + a.hostname;
    else return "";
    a.port && (b += ":" + a.port);
    return b;
  };
  d.buildAuthority = function (a) {
    return d.buildUserinfo(a) + d.buildHost(a);
  };
  d.buildUserinfo = function (a) {
    var b = "";
    a.username && (b += d.encode(a.username));
    a.password && (b += ":" + d.encode(a.password));
    b && (b += "@");
    return b;
  };
  d.buildQuery = function (a, b, c) {
    var e = "",
      f,
      g;
    for (f in a)
      if (l.call(a, f) && f)
        if (k(a[f])) {
          var n = {};
          var h = 0;
          for (g = a[f].length; h < g; h++)
            void 0 !== a[f][h] &&
              void 0 === n[a[f][h] + ""] &&
              ((e += "&" + d.buildQueryParameter(f, a[f][h], c)),
              !0 !== b && (n[a[f][h] + ""] = !0));
        } else
          void 0 !== a[f] && (e += "&" + d.buildQueryParameter(f, a[f], c));
    return e.substring(1);
  };
  d.buildQueryParameter = function (a, b, c) {
    return d.encodeQuery(a, c) + (null !== b ? "=" + d.encodeQuery(b, c) : "");
  };
  d.addQuery = function (a, b, c) {
    if ("object" === typeof b)
      for (var e in b) l.call(b, e) && d.addQuery(a, e, b[e]);
    else if ("string" === typeof b)
      void 0 === a[b]
        ? (a[b] = c)
        : ("string" === typeof a[b] && (a[b] = [a[b]]),
          k(c) || (c = [c]),
          (a[b] = (a[b] || []).concat(c)));
    else
      throw new TypeError(
        "URI.addQuery() accepts an object, string as the name parameter"
      );
  };
  d.setQuery = function (a, b, c) {
    if ("object" === typeof b)
      for (var e in b) l.call(b, e) && d.setQuery(a, e, b[e]);
    else if ("string" === typeof b) a[b] = void 0 === c ? null : c;
    else
      throw new TypeError(
        "URI.setQuery() accepts an object, string as the name parameter"
      );
  };
  d.removeQuery = function (a, b, c) {
    var e;
    if (k(b)) for (c = 0, e = b.length; c < e; c++) a[b[c]] = void 0;
    else if ("RegExp" === u(b)) for (e in a) b.test(e) && (a[e] = void 0);
    else if ("object" === typeof b)
      for (e in b) l.call(b, e) && d.removeQuery(a, e, b[e]);
    else if ("string" === typeof b)
      void 0 !== c
        ? "RegExp" === u(c)
          ? !k(a[b]) && c.test(a[b])
            ? (a[b] = void 0)
            : (a[b] = C(a[b], c))
          : a[b] !== String(c) || (k(c) && 1 !== c.length)
          ? k(a[b]) && (a[b] = C(a[b], c))
          : (a[b] = void 0)
        : (a[b] = void 0);
    else
      throw new TypeError(
        "URI.removeQuery() accepts an object, string, RegExp as the first parameter"
      );
  };
  d.hasQuery = function (a, b, c, e) {
    switch (u(b)) {
      case "String":
        break;
      case "RegExp":
        for (var f in a)
          if (
            l.call(a, f) &&
            b.test(f) &&
            (void 0 === c || d.hasQuery(a, f, c))
          )
            return !0;
        return !1;
      case "Object":
        for (var g in b) if (l.call(b, g) && !d.hasQuery(a, g, b[g])) return !1;
        return !0;
      default:
        throw new TypeError(
          "URI.hasQuery() accepts a string, regular expression or object as the name parameter"
        );
    }
    switch (u(c)) {
      case "Undefined":
        return b in a;
      case "Boolean":
        return (a = !(k(a[b]) ? !a[b].length : !a[b])), c === a;
      case "Function":
        return !!c(a[b], b, a);
      case "Array":
        return k(a[b]) ? (e ? w : D)(a[b], c) : !1;
      case "RegExp":
        return k(a[b]) ? (e ? w(a[b], c) : !1) : !(!a[b] || !a[b].match(c));
      case "Number":
        c = String(c);
      case "String":
        return k(a[b]) ? (e ? w(a[b], c) : !1) : a[b] === c;
      default:
        throw new TypeError(
          "URI.hasQuery() accepts undefined, boolean, string, number, RegExp, Function as the value parameter"
        );
    }
  };
  d.joinPaths = function () {
    for (var a = [], b = [], c = 0, e = 0; e < arguments.length; e++) {
      var f = new d(arguments[e]);
      a.push(f);
      f = f.segment();
      for (var g = 0; g < f.length; g++)
        "string" === typeof f[g] && b.push(f[g]), f[g] && c++;
    }
    if (!b.length || !c) return new d("");
    b = new d("").segment(b);
    ("" !== a[0].path() && "/" !== a[0].path().slice(0, 1)) ||
      b.path("/" + b.path());
    return b.normalize();
  };
  d.commonPath = function (a, b) {
    var c = Math.min(a.length, b.length),
      d;
    for (d = 0; d < c; d++)
      if (a.charAt(d) !== b.charAt(d)) {
        d--;
        break;
      }
    if (1 > d)
      return a.charAt(0) === b.charAt(0) && "/" === a.charAt(0) ? "/" : "";
    if ("/" !== a.charAt(d) || "/" !== b.charAt(d))
      d = a.substring(0, d).lastIndexOf("/");
    return a.substring(0, d + 1);
  };
  d.withinString = function (a, b, c) {
    c || (c = {});
    var e = c.start || d.findUri.start,
      f = c.end || d.findUri.end,
      g = c.trim || d.findUri.trim,
      n = c.parens || d.findUri.parens,
      k = /[a-z0-9-]=["']?$/i;
    for (e.lastIndex = 0; ; ) {
      var h = e.exec(a);
      if (!h) break;
      var m = h.index;
      if (c.ignoreHtml) {
        var q = a.slice(Math.max(m - 3, 0), m);
        if (q && k.test(q)) continue;
      }
      var l = m + a.slice(m).search(f);
      q = a.slice(m, l);
      for (l = -1; ; ) {
        var p = n.exec(q);
        if (!p) break;
        l = Math.max(l, p.index + p[0].length);
      }
      q = -1 < l ? q.slice(0, l) + q.slice(l).replace(g, "") : q.replace(g, "");
      q.length <= h[0].length ||
        (c.ignore && c.ignore.test(q)) ||
        ((l = m + q.length),
        (h = b(q, m, l, a)),
        void 0 === h
          ? (e.lastIndex = l)
          : ((h = String(h)),
            (a = a.slice(0, m) + h + a.slice(l)),
            (e.lastIndex = m + h.length)));
    }
    e.lastIndex = 0;
    return a;
  };
  d.ensureValidHostname = function (a, b) {
    var c = !!a,
      e = !1;
    b && (e = w(d.hostProtocols, b));
    if (e && !c)
      throw new TypeError("Hostname cannot be empty, if protocol is " + b);
    if (a && a.match(d.invalid_hostname_characters)) {
      if (!m)
        throw new TypeError(
          'Hostname "' +
            a +
            '" contains characters other than [A-Z0-9.-:_] and Punycode.js is not available'
        );
      if (m.toASCII(a).match(d.invalid_hostname_characters))
        throw new TypeError(
          'Hostname "' + a + '" contains characters other than [A-Z0-9.-:_]'
        );
    }
  };
  d.ensureValidPort = function (a) {
    if (a) {
      var b = Number(a);
      if (!(/^[0-9]+$/.test(b) && 0 < b && 65536 > b))
        throw new TypeError('Port "' + a + '" is not a valid port');
    }
  };
  d.noConflict = function (a) {
    if (a)
      return (
        (a = { URI: this.noConflict() }),
        h.URITemplate &&
          "function" === typeof h.URITemplate.noConflict &&
          (a.URITemplate = h.URITemplate.noConflict()),
        h.IPv6 &&
          "function" === typeof h.IPv6.noConflict &&
          (a.IPv6 = h.IPv6.noConflict()),
        h.SecondLevelDomains &&
          "function" === typeof h.SecondLevelDomains.noConflict &&
          (a.SecondLevelDomains = h.SecondLevelDomains.noConflict()),
        a
      );
    h.URI === this && (h.URI = G);
    return this;
  };
  g.build = function (a) {
    if (!0 === a) this._deferred_build = !0;
    else if (void 0 === a || this._deferred_build)
      (this._string = d.build(this._parts)), (this._deferred_build = !1);
    return this;
  };
  g.clone = function () {
    return new d(this);
  };
  g.valueOf = g.toString = function () {
    return this.build(!1)._string;
  };
  g.protocol = x("protocol");
  g.username = x("username");
  g.password = x("password");
  g.hostname = x("hostname");
  g.port = x("port");
  g.query = E("query", "?");
  g.fragment = E("fragment", "#");
  g.search = function (a, b) {
    var c = this.query(a, b);
    return "string" === typeof c && c.length ? "?" + c : c;
  };
  g.hash = function (a, b) {
    var c = this.fragment(a, b);
    return "string" === typeof c && c.length ? "#" + c : c;
  };
  g.pathname = function (a, b) {
    if (void 0 === a || !0 === a) {
      var c = this._parts.path || (this._parts.hostname ? "/" : "");
      return a ? (this._parts.urn ? d.decodeUrnPath : d.decodePath)(c) : c;
    }
    this._parts.path = this._parts.urn
      ? a
        ? d.recodeUrnPath(a)
        : ""
      : a
      ? d.recodePath(a)
      : "/";
    this.build(!b);
    return this;
  };
  g.path = g.pathname;
  g.href = function (a, b) {
    var c;
    if (void 0 === a) return this.toString();
    this._string = "";
    this._parts = d._parts();
    var e = a instanceof d,
      f = "object" === typeof a && (a.hostname || a.path || a.pathname);
    a.nodeName && ((f = d.getDomAttribute(a)), (a = a[f] || ""), (f = !1));
    !e && f && void 0 !== a.pathname && (a = a.toString());
    if ("string" === typeof a || a instanceof String)
      this._parts = d.parse(String(a), this._parts);
    else if (e || f) {
      e = e ? a._parts : a;
      for (c in e)
        "query" !== c && l.call(this._parts, c) && (this._parts[c] = e[c]);
      e.query && this.query(e.query, !1);
    } else throw new TypeError("invalid input");
    this.build(!b);
    return this;
  };
  g.is = function (a) {
    var b = !1,
      c = !1,
      e = !1,
      f = !1,
      g = !1,
      h = !1,
      k = !1,
      l = !this._parts.urn;
    this._parts.hostname &&
      ((l = !1),
      (c = d.ip4_expression.test(this._parts.hostname)),
      (e = d.ip6_expression.test(this._parts.hostname)),
      (b = c || e),
      (g = (f = !b) && t && t.has(this._parts.hostname)),
      (h = f && d.idn_expression.test(this._parts.hostname)),
      (k = f && d.punycode_expression.test(this._parts.hostname)));
    switch (a.toLowerCase()) {
      case "relative":
        return l;
      case "absolute":
        return !l;
      case "domain":
      case "name":
        return f;
      case "sld":
        return g;
      case "ip":
        return b;
      case "ip4":
      case "ipv4":
      case "inet4":
        return c;
      case "ip6":
      case "ipv6":
      case "inet6":
        return e;
      case "idn":
        return h;
      case "url":
        return !this._parts.urn;
      case "urn":
        return !!this._parts.urn;
      case "punycode":
        return k;
    }
    return null;
  };
  var H = g.protocol,
    I = g.port,
    J = g.hostname;
  g.protocol = function (a, b) {
    if (
      a &&
      ((a = a.replace(/:(\/\/)?$/, "")), !a.match(d.protocol_expression))
    )
      throw new TypeError(
        'Protocol "' +
          a +
          "\" contains characters other than [A-Z0-9.+-] or doesn't start with [A-Z]"
      );
    return H.call(this, a, b);
  };
  g.scheme = g.protocol;
  g.port = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    void 0 !== a &&
      (0 === a && (a = null),
      a &&
        ((a += ""),
        ":" === a.charAt(0) && (a = a.substring(1)),
        d.ensureValidPort(a)));
    return I.call(this, a, b);
  };
  g.hostname = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    if (void 0 !== a) {
      var c = { preventInvalidHostname: this._parts.preventInvalidHostname };
      if ("/" !== d.parseHost(a, c))
        throw new TypeError(
          'Hostname "' + a + '" contains characters other than [A-Z0-9.-]'
        );
      a = c.hostname;
      this._parts.preventInvalidHostname &&
        d.ensureValidHostname(a, this._parts.protocol);
    }
    return J.call(this, a, b);
  };
  g.origin = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    if (void 0 === a) {
      var c = this.protocol();
      return this.authority() ? (c ? c + "://" : "") + this.authority() : "";
    }
    c = d(a);
    this.protocol(c.protocol()).authority(c.authority()).build(!b);
    return this;
  };
  g.host = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    if (void 0 === a)
      return this._parts.hostname ? d.buildHost(this._parts) : "";
    if ("/" !== d.parseHost(a, this._parts))
      throw new TypeError(
        'Hostname "' + a + '" contains characters other than [A-Z0-9.-]'
      );
    this.build(!b);
    return this;
  };
  g.authority = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    if (void 0 === a)
      return this._parts.hostname ? d.buildAuthority(this._parts) : "";
    if ("/" !== d.parseAuthority(a, this._parts))
      throw new TypeError(
        'Hostname "' + a + '" contains characters other than [A-Z0-9.-]'
      );
    this.build(!b);
    return this;
  };
  g.userinfo = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    if (void 0 === a) {
      var c = d.buildUserinfo(this._parts);
      return c ? c.substring(0, c.length - 1) : c;
    }
    "@" !== a[a.length - 1] && (a += "@");
    d.parseUserinfo(a, this._parts);
    this.build(!b);
    return this;
  };
  g.resource = function (a, b) {
    if (void 0 === a) return this.path() + this.search() + this.hash();
    var c = d.parse(a);
    this._parts.path = c.path;
    this._parts.query = c.query;
    this._parts.fragment = c.fragment;
    this.build(!b);
    return this;
  };
  g.subdomain = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    if (void 0 === a) {
      if (!this._parts.hostname || this.is("IP")) return "";
      var c = this._parts.hostname.length - this.domain().length - 1;
      return this._parts.hostname.substring(0, c) || "";
    }
    c = this._parts.hostname.length - this.domain().length;
    c = this._parts.hostname.substring(0, c);
    c = new RegExp("^" + p(c));
    a && "." !== a.charAt(a.length - 1) && (a += ".");
    if (-1 !== a.indexOf(":"))
      throw new TypeError("Domains cannot contain colons");
    a && d.ensureValidHostname(a, this._parts.protocol);
    this._parts.hostname = this._parts.hostname.replace(c, a);
    this.build(!b);
    return this;
  };
  g.domain = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    "boolean" === typeof a && ((b = a), (a = void 0));
    if (void 0 === a) {
      if (!this._parts.hostname || this.is("IP")) return "";
      var c = this._parts.hostname.match(/\./g);
      if (c && 2 > c.length) return this._parts.hostname;
      c = this._parts.hostname.length - this.tld(b).length - 1;
      c = this._parts.hostname.lastIndexOf(".", c - 1) + 1;
      return this._parts.hostname.substring(c) || "";
    }
    if (!a) throw new TypeError("cannot set domain empty");
    if (-1 !== a.indexOf(":"))
      throw new TypeError("Domains cannot contain colons");
    d.ensureValidHostname(a, this._parts.protocol);
    !this._parts.hostname || this.is("IP")
      ? (this._parts.hostname = a)
      : ((c = new RegExp(p(this.domain()) + "$")),
        (this._parts.hostname = this._parts.hostname.replace(c, a)));
    this.build(!b);
    return this;
  };
  g.tld = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    "boolean" === typeof a && ((b = a), (a = void 0));
    if (void 0 === a) {
      if (!this._parts.hostname || this.is("IP")) return "";
      var c = this._parts.hostname.lastIndexOf(".");
      c = this._parts.hostname.substring(c + 1);
      return !0 !== b && t && t.list[c.toLowerCase()]
        ? t.get(this._parts.hostname) || c
        : c;
    }
    if (a)
      if (a.match(/[^a-zA-Z0-9-]/))
        if (t && t.is(a))
          (c = new RegExp(p(this.tld()) + "$")),
            (this._parts.hostname = this._parts.hostname.replace(c, a));
        else
          throw new TypeError(
            'TLD "' + a + '" contains characters other than [A-Z0-9]'
          );
      else {
        if (!this._parts.hostname || this.is("IP"))
          throw new ReferenceError("cannot set TLD on non-domain host");
        c = new RegExp(p(this.tld()) + "$");
        this._parts.hostname = this._parts.hostname.replace(c, a);
      }
    else throw new TypeError("cannot set TLD empty");
    this.build(!b);
    return this;
  };
  g.directory = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    if (void 0 === a || !0 === a) {
      if (!this._parts.path && !this._parts.hostname) return "";
      if ("/" === this._parts.path) return "/";
      var c = this._parts.path.length - this.filename().length - 1;
      c = this._parts.path.substring(0, c) || (this._parts.hostname ? "/" : "");
      return a ? d.decodePath(c) : c;
    }
    c = this._parts.path.length - this.filename().length;
    c = this._parts.path.substring(0, c);
    c = new RegExp("^" + p(c));
    this.is("relative") ||
      (a || (a = "/"), "/" !== a.charAt(0) && (a = "/" + a));
    a && "/" !== a.charAt(a.length - 1) && (a += "/");
    a = d.recodePath(a);
    this._parts.path = this._parts.path.replace(c, a);
    this.build(!b);
    return this;
  };
  g.filename = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    if ("string" !== typeof a) {
      if (!this._parts.path || "/" === this._parts.path) return "";
      var c = this._parts.path.lastIndexOf("/");
      c = this._parts.path.substring(c + 1);
      return a ? d.decodePathSegment(c) : c;
    }
    c = !1;
    "/" === a.charAt(0) && (a = a.substring(1));
    a.match(/\.?\//) && (c = !0);
    var e = new RegExp(p(this.filename()) + "$");
    a = d.recodePath(a);
    this._parts.path = this._parts.path.replace(e, a);
    c ? this.normalizePath(b) : this.build(!b);
    return this;
  };
  g.suffix = function (a, b) {
    if (this._parts.urn) return void 0 === a ? "" : this;
    if (void 0 === a || !0 === a) {
      if (!this._parts.path || "/" === this._parts.path) return "";
      var c = this.filename(),
        e = c.lastIndexOf(".");
      if (-1 === e) return "";
      c = c.substring(e + 1);
      c = /^[a-z0-9%]+$/i.test(c) ? c : "";
      return a ? d.decodePathSegment(c) : c;
    }
    "." === a.charAt(0) && (a = a.substring(1));
    if ((c = this.suffix()))
      e = a ? new RegExp(p(c) + "$") : new RegExp(p("." + c) + "$");
    else {
      if (!a) return this;
      this._parts.path += "." + d.recodePath(a);
    }
    e &&
      ((a = d.recodePath(a)),
      (this._parts.path = this._parts.path.replace(e, a)));
    this.build(!b);
    return this;
  };
  g.segment = function (a, b, c) {
    var d = this._parts.urn ? ":" : "/",
      f = this.path(),
      g = "/" === f.substring(0, 1);
    f = f.split(d);
    void 0 !== a && "number" !== typeof a && ((c = b), (b = a), (a = void 0));
    if (void 0 !== a && "number" !== typeof a)
      throw Error('Bad segment "' + a + '", must be 0-based integer');
    g && f.shift();
    0 > a && (a = Math.max(f.length + a, 0));
    if (void 0 === b) return void 0 === a ? f : f[a];
    if (null === a || void 0 === f[a])
      if (k(b)) {
        f = [];
        a = 0;
        for (var h = b.length; a < h; a++)
          if (b[a].length || (f.length && f[f.length - 1].length))
            f.length && !f[f.length - 1].length && f.pop(), f.push(z(b[a]));
      } else {
        if (b || "string" === typeof b)
          (b = z(b)),
            "" === f[f.length - 1] ? (f[f.length - 1] = b) : f.push(b);
      }
    else b ? (f[a] = z(b)) : f.splice(a, 1);
    g && f.unshift("");
    return this.path(f.join(d), c);
  };
  g.segmentCoded = function (a, b, c) {
    var e;
    "number" !== typeof a && ((c = b), (b = a), (a = void 0));
    if (void 0 === b) {
      a = this.segment(a, b, c);
      if (k(a)) {
        var f = 0;
        for (e = a.length; f < e; f++) a[f] = d.decode(a[f]);
      } else a = void 0 !== a ? d.decode(a) : void 0;
      return a;
    }
    if (k(b)) for (f = 0, e = b.length; f < e; f++) b[f] = d.encode(b[f]);
    else b = "string" === typeof b || b instanceof String ? d.encode(b) : b;
    return this.segment(a, b, c);
  };
  var K = g.query;
  g.query = function (a, b) {
    if (!0 === a)
      return d.parseQuery(this._parts.query, this._parts.escapeQuerySpace);
    if ("function" === typeof a) {
      var c = d.parseQuery(this._parts.query, this._parts.escapeQuerySpace),
        e = a.call(this, c);
      this._parts.query = d.buildQuery(
        e || c,
        this._parts.duplicateQueryParameters,
        this._parts.escapeQuerySpace
      );
      this.build(!b);
      return this;
    }
    return void 0 !== a && "string" !== typeof a
      ? ((this._parts.query = d.buildQuery(
          a,
          this._parts.duplicateQueryParameters,
          this._parts.escapeQuerySpace
        )),
        this.build(!b),
        this)
      : K.call(this, a, b);
  };
  g.setQuery = function (a, b, c) {
    var e = d.parseQuery(this._parts.query, this._parts.escapeQuerySpace);
    if ("string" === typeof a || a instanceof String)
      e[a] = void 0 !== b ? b : null;
    else if ("object" === typeof a)
      for (var f in a) l.call(a, f) && (e[f] = a[f]);
    else
      throw new TypeError(
        "URI.addQuery() accepts an object, string as the name parameter"
      );
    this._parts.query = d.buildQuery(
      e,
      this._parts.duplicateQueryParameters,
      this._parts.escapeQuerySpace
    );
    "string" !== typeof a && (c = b);
    this.build(!c);
    return this;
  };
  g.addQuery = function (a, b, c) {
    var e = d.parseQuery(this._parts.query, this._parts.escapeQuerySpace);
    d.addQuery(e, a, void 0 === b ? null : b);
    this._parts.query = d.buildQuery(
      e,
      this._parts.duplicateQueryParameters,
      this._parts.escapeQuerySpace
    );
    "string" !== typeof a && (c = b);
    this.build(!c);
    return this;
  };
  g.removeQuery = function (a, b, c) {
    var e = d.parseQuery(this._parts.query, this._parts.escapeQuerySpace);
    d.removeQuery(e, a, b);
    this._parts.query = d.buildQuery(
      e,
      this._parts.duplicateQueryParameters,
      this._parts.escapeQuerySpace
    );
    "string" !== typeof a && (c = b);
    this.build(!c);
    return this;
  };
  g.hasQuery = function (a, b, c) {
    var e = d.parseQuery(this._parts.query, this._parts.escapeQuerySpace);
    return d.hasQuery(e, a, b, c);
  };
  g.setSearch = g.setQuery;
  g.addSearch = g.addQuery;
  g.removeSearch = g.removeQuery;
  g.hasSearch = g.hasQuery;
  g.normalize = function () {
    return this._parts.urn
      ? this.normalizeProtocol(!1)
          .normalizePath(!1)
          .normalizeQuery(!1)
          .normalizeFragment(!1)
          .build()
      : this.normalizeProtocol(!1)
          .normalizeHostname(!1)
          .normalizePort(!1)
          .normalizePath(!1)
          .normalizeQuery(!1)
          .normalizeFragment(!1)
          .build();
  };
  g.normalizeProtocol = function (a) {
    "string" === typeof this._parts.protocol &&
      ((this._parts.protocol = this._parts.protocol.toLowerCase()),
      this.build(!a));
    return this;
  };
  g.normalizeHostname = function (a) {
    this._parts.hostname &&
      (this.is("IDN") && m
        ? (this._parts.hostname = m.toASCII(this._parts.hostname))
        : this.is("IPv6") &&
          v &&
          (this._parts.hostname = v.best(this._parts.hostname)),
      (this._parts.hostname = this._parts.hostname.toLowerCase()),
      this.build(!a));
    return this;
  };
  g.normalizePort = function (a) {
    "string" === typeof this._parts.protocol &&
      this._parts.port === d.defaultPorts[this._parts.protocol] &&
      ((this._parts.port = null), this.build(!a));
    return this;
  };
  g.normalizePath = function (a) {
    var b = this._parts.path;
    if (!b) return this;
    if (this._parts.urn)
      return (
        (this._parts.path = d.recodeUrnPath(this._parts.path)),
        this.build(!a),
        this
      );
    if ("/" === this._parts.path) return this;
    b = d.recodePath(b);
    var c = "";
    if ("/" !== b.charAt(0)) {
      var e = !0;
      b = "/" + b;
    }
    if ("/.." === b.slice(-3) || "/." === b.slice(-2)) b += "/";
    b = b.replace(/(\/(\.\/)+)|(\/\.$)/g, "/").replace(/\/{2,}/g, "/");
    e && (c = b.substring(1).match(/^(\.\.\/)+/) || "") && (c = c[0]);
    for (;;) {
      var f = b.search(/\/\.\.(\/|$)/);
      if (-1 === f) break;
      else if (0 === f) {
        b = b.substring(3);
        continue;
      }
      var g = b.substring(0, f).lastIndexOf("/");
      -1 === g && (g = f);
      b = b.substring(0, g) + b.substring(f + 3);
    }
    e && this.is("relative") && (b = c + b.substring(1));
    this._parts.path = b;
    this.build(!a);
    return this;
  };
  g.normalizePathname = g.normalizePath;
  g.normalizeQuery = function (a) {
    "string" === typeof this._parts.query &&
      (this._parts.query.length
        ? this.query(
            d.parseQuery(this._parts.query, this._parts.escapeQuerySpace)
          )
        : (this._parts.query = null),
      this.build(!a));
    return this;
  };
  g.normalizeFragment = function (a) {
    this._parts.fragment || ((this._parts.fragment = null), this.build(!a));
    return this;
  };
  g.normalizeSearch = g.normalizeQuery;
  g.normalizeHash = g.normalizeFragment;
  g.iso8859 = function () {
    var a = d.encode,
      b = d.decode;
    d.encode = escape;
    d.decode = decodeURIComponent;
    try {
      this.normalize();
    } finally {
      (d.encode = a), (d.decode = b);
    }
    return this;
  };
  g.unicode = function () {
    var a = d.encode,
      b = d.decode;
    d.encode = A;
    d.decode = unescape;
    try {
      this.normalize();
    } finally {
      (d.encode = a), (d.decode = b);
    }
    return this;
  };
  g.readable = function () {
    var a = this.clone();
    a.username("").password("").normalize();
    var b = "";
    a._parts.protocol && (b += a._parts.protocol + "://");
    a._parts.hostname &&
      (a.is("punycode") && m
        ? ((b += m.toUnicode(a._parts.hostname)),
          a._parts.port && (b += ":" + a._parts.port))
        : (b += a.host()));
    a._parts.hostname &&
      a._parts.path &&
      "/" !== a._parts.path.charAt(0) &&
      (b += "/");
    b += a.path(!0);
    if (a._parts.query) {
      for (
        var c = "", e = 0, f = a._parts.query.split("&"), g = f.length;
        e < g;
        e++
      ) {
        var h = (f[e] || "").split("=");
        c +=
          "&" +
          d
            .decodeQuery(h[0], this._parts.escapeQuerySpace)
            .replace(/&/g, "%26");
        void 0 !== h[1] &&
          (c +=
            "=" +
            d
              .decodeQuery(h[1], this._parts.escapeQuerySpace)
              .replace(/&/g, "%26"));
      }
      b += "?" + c.substring(1);
    }
    return (b += d.decodeQuery(a.hash(), !0));
  };
  g.absoluteTo = function (a) {
    var b = this.clone(),
      c = ["protocol", "username", "password", "hostname", "port"],
      e,
      f;
    if (this._parts.urn)
      throw Error(
        "URNs do not have any generally defined hierarchical components"
      );
    a instanceof d || (a = new d(a));
    if (b._parts.protocol) return b;
    b._parts.protocol = a._parts.protocol;
    if (this._parts.hostname) return b;
    for (e = 0; (f = c[e]); e++) b._parts[f] = a._parts[f];
    b._parts.path
      ? (".." === b._parts.path.substring(-2) && (b._parts.path += "/"),
        "/" !== b.path().charAt(0) &&
          ((c = (c = a.directory())
            ? c
            : 0 === a.path().indexOf("/")
            ? "/"
            : ""),
          (b._parts.path = (c ? c + "/" : "") + b._parts.path),
          b.normalizePath()))
      : ((b._parts.path = a._parts.path),
        b._parts.query || (b._parts.query = a._parts.query));
    b.build();
    return b;
  };
  g.relativeTo = function (a) {
    var b = this.clone().normalize();
    if (b._parts.urn)
      throw Error(
        "URNs do not have any generally defined hierarchical components"
      );
    a = new d(a).normalize();
    var c = b._parts;
    var e = a._parts;
    var f = b.path();
    a = a.path();
    if ("/" !== f.charAt(0)) throw Error("URI is already relative");
    if ("/" !== a.charAt(0))
      throw Error("Cannot calculate a URI relative to another relative URI");
    c.protocol === e.protocol && (c.protocol = null);
    if (
      c.username === e.username &&
      c.password === e.password &&
      null === c.protocol &&
      null === c.username &&
      null === c.password &&
      c.hostname === e.hostname &&
      c.port === e.port
    )
      (c.hostname = null), (c.port = null);
    else return b.build();
    if (f === a) return (c.path = ""), b.build();
    f = d.commonPath(f, a);
    if (!f) return b.build();
    e = e.path
      .substring(f.length)
      .replace(/[^\/]*$/, "")
      .replace(/.*?\//g, "../");
    c.path = e + c.path.substring(f.length) || "./";
    return b.build();
  };
  g.equals = function (a) {
    var b = this.clone(),
      c = new d(a);
    a = {};
    var e;
    b.normalize();
    c.normalize();
    if (b.toString() === c.toString()) return !0;
    var f = b.query();
    var g = c.query();
    b.query("");
    c.query("");
    if (b.toString() !== c.toString() || f.length !== g.length) return !1;
    b = d.parseQuery(f, this._parts.escapeQuerySpace);
    g = d.parseQuery(g, this._parts.escapeQuerySpace);
    for (e in b)
      if (l.call(b, e)) {
        if (!k(b[e])) {
          if (b[e] !== g[e]) return !1;
        } else if (!D(b[e], g[e])) return !1;
        a[e] = !0;
      }
    for (e in g) if (l.call(g, e) && !a[e]) return !1;
    return !0;
  };
  g.preventInvalidHostname = function (a) {
    this._parts.preventInvalidHostname = !!a;
    return this;
  };
  g.duplicateQueryParameters = function (a) {
    this._parts.duplicateQueryParameters = !!a;
    return this;
  };
  g.escapeQuerySpace = function (a) {
    this._parts.escapeQuerySpace = !!a;
    return this;
  };
  return d;
});
