export default {
  install(Vue, options) {
    Vue.prototype.isEmail = function(s) {
    	return /^((([a-z]|\d|[!#\$%&'\*\+\-\/=\?\^_`{\|}~]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])+(\.([a-z]|\d|[!#\$%&'\*\+\-\/=\?\^_`{\|}~]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])+)*)|((\x22)((((\x20|\x09)*(\x0d\x0a))?(\x20|\x09)+)?(([\x01-\x08\x0b\x0c\x0e-\x1f\x7f]|\x21|[\x23-\x5b]|[\x5d-\x7e]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(\\([\x01-\x09\x0b\x0c\x0d-\x7f]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF]))))*(((\x20|\x09)*(\x0d\x0a))?(\x20|\x09)+)?(\x22)))@((([a-z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(([a-z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])([a-z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])*([a-z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])))\.)+(([a-z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(([a-z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])([a-z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])*([a-z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])))\.?$/i.test(s.trim());
    },

    Vue.prototype.formatDate = function(date, fmt) {
      date = date || new Date();
      date = typeof date == 'number' ? new Date(date) : date;
      fmt = fmt || 'yyyy-MM-dd HH:mm:ss';
      let obj = {
        'y': date.getFullYear(), // 年份，注意必须用getFullYear
        'M': date.getMonth() + 1, // 月份，注意是从0-11
        'd': date.getDate(), // 日期
        'q': Math.floor((date.getMonth() + 3) / 3), // 季度
        'w': date.getDay(), // 星期，注意是0-6
        'H': date.getHours(), // 24小时制
        'h': date.getHours() % 12 == 0 ? 12 : date.getHours() % 12, // 12小时制
        'm': date.getMinutes(), // 分钟
        's': date.getSeconds(), // 秒
        'S': date.getMilliseconds() // 毫秒
      };
      
      let week = [ '天', '一', '二', '三', '四', '五', '六' ];
      for (let i in obj) {
        fmt = fmt.replace(new RegExp(i+'+', 'g'), function(m) {
            let val = obj[i] + '';
            if(i == 'w') return (m.length > 2 ? '星期' : '周') + week[val];
            for(let j = 0, len = val.length; j < m.length - len; j++) val = '0' + val;
            return m.length == 1 ? val : val.substring(val.length - m.length);
        });
      }

      return fmt;
    },

    Vue.prototype.parseDate = function(str, fmt) {
      fmt = fmt || 'yyyy-MM-dd';
      let obj = { y: 0, M: 1, d: 0, H: 0, h: 0, m: 0, s: 0, S: 0 };
      fmt.replace(/([^yMdHmsS]*?)(([yMdHmsS])\3*)([^yMdHmsS]*?)/g, function(m, $1, $2, $3, $4, idx, old) {
        str = str.replace(new RegExp($1+'(\\d{'+$2.length+'})'+$4), function(_m, _$1) {
          obj[$3] = parseInt(_$1);
          return '';
        });

        return '';
      });

      obj.M--; // 月份是从0开始的，所以要减去1
      let date = new Date(obj.y, obj.M, obj.d, obj.H, obj.m, obj.s);
      if (obj.S !== 0) date.setMilliseconds(obj.S); // 如果设置了毫秒
      return date;
    },

    Vue.prototype.moduleName = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      return Vue.prototype.constants.MODULE_MAP[id].Name;
    },

    Vue.prototype.operationName = function(moduleId, id) {
      if (!id || id.length == 0) {
        return '';
      }

      let list = Vue.prototype.constants.OPERATION_MAP[moduleId];
      for (let e of list) {
        if (id == e.Id) {
          return e.Name;
        }
      }

      return '';
    },

    Vue.prototype.sshOpName = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      let list = Vue.prototype.constants.SSH_OP_LIST;
      for (let e of list) {
        if (id == e.Id) {
          return e.Name;
        }
      }

      return '';
    },

    Vue.prototype.applicationStatus = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      let m = Vue.prototype.constants.APPLICATION_STATUS_MAP;
      let s = m[id];
      return s ? s : m["*"];
    },

    Vue.prototype.applicationClass = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      let m = Vue.prototype.constants.APPLICATION_CLASS_MAP;
      let s = m[id];
      return s ? s : m["*"];
    },

    Vue.prototype.containerStatus = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      let m = Vue.prototype.constants.CONTAINER_STATUS_MAP;
      let s = m[id];
      return s ? s : m["*"];
    },

    Vue.prototype.containerClass = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      let m = Vue.prototype.constants.CONTAINER_CLASS_MAP;
      let s = m[id];
      return s ? s : m["*"];
    },
    
    Vue.prototype.validateForm = function(refPrefix) {
      let result = true;
    	Object.keys(this.$refs).forEach(k => {
        if (!refPrefix || k.indexOf(refPrefix) == 0) {
          let e = this.$refs[k];
          if (Array.isArray(e)) {
            e = e[0];
          }

          if (e && e.validate) {
            e.validate();
            if (e.errorBucket && e.errorBucket.length > 0) {
              result = false;
            }
          }
        }
      });

      return result;
    }
  }
}