/**
 60 * 功能： 判断此浏览器是在PC 端， 还是移动端。
 61 *
 62 * 返回值： false ， 说明当前操作系统是移动端；
 63 *  true ， 说明当前的操作系统是PC 端。
 64 */
function IsPC() {
  var userAgentInfo = navigator.userAgent;
  var Agents = ["Android", "iPhone", "SymbianOS", "Windows Phone", "iPad", "iPod"];
  var flag = true;

  for (var v = 0; v < Agents.length; v++) {
    if (userAgentInfo.indexOf(Agents[v]) > 0) {
      flag = false;
      break;
    }
  }
  return flag;
}

/**
 81 * 功能： 判断是Android 端还是iOS 端。
 82 *
 83 * 返回值： true ， 说明是Android 端；
 84 * false ， 说明是iOS 端。
 85 */
function IsAndroid() {
  var u = navigator.userAgent, app = navigator.appVersion;
  var isAndroid = u.indexOf('Android ') > -1 || u.indexOf('Linux ') > -1;
  var isIOS = !!u.match(/\(i[^;]+;( U;)? CPU.+Mac OS X/);
  if (isAndroid) {
    // 这个是Android 系统
    return true;
  }

  if (isIOS) {
    // 这个是iOS 系统
    return false;
  }
}


/**
 102 * 功能： 从url 中获取指定的域值
 103 *
 104 * 返回值： 指定的域值或false
 105 */
function getQueryVariable(variable) {
  var query = window.location.search.substring(1);
  var vars = query.split("&");
  for (var i = 0; i < vars.length; i++) {
    var pair = vars[i].split("=");
    if (pair [0] == variable) {
      return pair [1];
    }
  }
  return false;
}

/**
 367 * 功能： 错误处理函数
 368 *
 369 * 返回值： 无
 370 */
function handleError(err) {
  console.error('Failed to get Media Stream!', err);
}

// 处理Offer 错误
function handleOfferError(err) {
  console.error('Failed to create offer:', err);
}


/**
 429 * 功能： 处理Answer 错误
 430 *
 431 * 返回值： 无
 432 */
function handleAnswerError(err) {
  console.error('Failed to create answer:', err);
}
