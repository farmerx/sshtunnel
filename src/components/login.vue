<template>
  <div class="login" :style="winSize">
    <el-form :model="loginForm" :rules="rules" :style="formOffset" ref="loginForm" label-position="left" label-width="100px" class="demo-ruleForm card-box loginform">
     <h3 class="title">
      <span>SSHTunel系统登录</span>                       
    </h3>
      <el-form-item label="用户" prop="user">
        <el-input type="input" v-model="loginForm.user" auto-complete="off"></el-input>
      </el-form-item>
      <el-form-item label="密码" prop="pass">
        <el-input type="password" v-model="loginForm.pass" auto-complete="off"></el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="submitForm('loginForm')">提交</el-button>
        <el-button @click="resetForm('loginForm')">重置</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
export default {
  name: 'login',
  data () {
    var validatePass = (rule, value, callback) => {
      if (value === '') {
        return callback(new Error('请输入密码'))
      } else if (value.length < 5) {
        return callback(new Error('密码长度不小于五个字符'))
      } else {
        callback()
      }
    }
    var validateUser = (rule, value, callback) => {
      if (value === '') {
        return callback(new Error('请输入用户名'))
      } else {
        if (value.length < 5) {
          return callback(new Error('用户名长度不小于五个字符'))
        } else {
          callback()
        }
      }
    }
    return {
      winSize: {
        width: '',
        height: ''
      },
      formOffset: {
        position: 'absolute',
        left: '',
        top: ''
      },
      loginForm: {
        pass: '',
        user: ''
      },
      rules: {
        pass: [
          { validator: validatePass, trigger: 'blur' }
        ],
        user: [
          { validator: validateUser, trigger: 'blur' }
        ]
      }
    }
  },
  methods: {
    setSize () {
      this.winSize.width = window.innerWidth + 'px'
      this.winSize.height = window.innerHeight + 'px'
      this.formOffset.left = (parseInt(this.winSize.width) / 2 - 175) + 'px'
      this.formOffset.top = (parseInt(this.winSize.height) / 2 - 178) + 'px'
    },
    submitForm (formName) {
      this.$refs[formName].validate((valid) => {
        if (valid) {
          var tmp = {'url': '/login.json', 'pass': this.loginForm.pass, 'user': this.loginForm.user, 'action': this.Action}
          this.$store.dispatch('CookieSetUserInfo', tmp)
        } else {
          return false
        }
      })
    },
    resetForm (formName) {
      this.$refs[formName].resetFields()
    }
  },
  computed: {
    ...mapGetters([
      'LoginError',
      'LoginFlag',
      'Action'
    ])
  },
  watch: {
    Action () {
      if (this.LoginFlag === false) {
        this.$message.error(this.LoginError)
      } else {
        this.$router.push({ path: '/' })
      }
    }
  },
  created () {
    this.setSize()
    if (this.$cookie.get('SSHTunnelUserInfo') !== null) {
      this.$router.push({path: '/'})
    }
    // function setSize2 () {
    //   this.winSize.width = window.innerWidth + 'px'
    //   this.winSize.height = window.innerHeight + 'px'
    //   this.formOffset.left = (parseInt(this.winSize.width) / 2 - 175) + 'px'
    //   this.formOffset.top = (parseInt(this.winSize.height) / 2 - 178) + 'px'
    // }
    // var a = this.setSize
    // window.onresize = function () {
    //   setSize2()
    // }
  }
}
</script>

<style scoped lang='less'>
  .login {
    background: #1F2D3D;
    .card-box {
        box-shadow            : 0 0px 8px 0 rgba(0, 0, 0, 0.06), 0 1px 0px 0 rgba(0, 0, 0, 0.02);
        -webkit-border-radius : 5px;
        border-radius         : 5px;
        -moz-border-radius    : 5px;
        background-clip       : padding-box;
        margin-bottom         : 20px;
        background-color      : #F9FAFC;
        border                : 2px solid #8492A6;
    }
    .title {
        margin      : 0px auto 40px auto;
        text-align  : center;
        color       : #505458;
        font-weight : normal;
        font-size   : 16px;

        span {
            cursor : pointer;
            &.active {
                font-weight : bold;
                font-size   : 18px;
            }
        }
    }
    .loginform {
        width   : 350px;
        padding : 35px 35px 15px 35px;
    }
  }
</style>]

