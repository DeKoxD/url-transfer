var id;
var url = "";
var serverAddr = window.location.protocol + "//" + window.location.host;

var lggr;

var elH2ID = document.getElementById("id");

var elButtonGetURL = document.getElementById("get-url-button");
var elAURL = document.getElementById("url-anchor")
var elDivQRCode = document.getElementById("qrcode")

const params = getURLParams();

if(params.logs && (params.logs !== "false" || params.logs !== "0")) {
	lggr = new logger(document.getElementById('log'));
} else {
	lggr = { log: function() {} };
}

function getParamsString(params) {
	const paramStrings = [];
	for(var key in params) {
		if(params[key]) paramStrings.push(key + '=' + params[key]);
	}
	return paramStrings.join('&');
}

var xhrURL = new XMLHttpRequest();
xhrURL.onload = function () {
	lggr.log(xhrURL.response)
	try {
		var response = JSON.parse(xhrURL.response)
		if(response.url) {
			url = response.url;
			elAURL.innerHTML = url;
			elAURL.href = url;
		} else {
			elAURL.innerHTML = "No URL found";
			elAURL.removeAttribute("href")
		}
		if(!id || !id.length) {
			if(response.id) {
				if(response.id.length) {
					id = response.id;
					elH2ID.innerText = "ID: " + id;
				}
			}
		}
		lggr.log(url);
	} catch (error) {
		lggr.log(error)
	}
}
xhrURL.onloadstart = function() {
	elAURL.innerHTML = "Fetching...";
	elAURL.removeAttribute("href")
}
xhrURL.onerror = function() {
	elAURL.innerHTML = "Error!";
	elAURL.removeAttribute("href")
}

elButtonGetURL.onclick = function() {
	xhrURL.open('GET', serverAddr + '/url/' + id);
	xhrURL.send();
}


// Register
var xhrRegister = new XMLHttpRequest();
xhrRegister.onload = function () {
	try {
		var response = JSON.parse(xhrRegister.response)
		if(response.id) {
			id = response.id
			elButtonGetURL.disabled = false;
			elDivQRCode.src = serverAddr + '/qrcode?url=' + serverAddr + '/send?id=' + id
			elH2ID.innerText = "ID: " + id;
			lggr.log(id);
		}
	} catch (error) {
		lggr.log(error)
	}
}
xhrRegister.open('GET', serverAddr + '/register');
xhrRegister.send();
