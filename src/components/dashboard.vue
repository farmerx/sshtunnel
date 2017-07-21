<template>
  <div class="dashboard">
    <el-menu theme="dark" :default-active="activeIndex" class="el-menu-demo" mode="horizontal" @select="handleSelect">
        <el-menu-item index="1">端口映射列表</el-menu-item>
        <!--<el-menu-item index="2">服务器信息</el-menu-item>
        <el-menu-item index="3">API接口</el-menu-item>
        <el-menu-item index="4">About</el-menu-item>-->
        <el-menu-item index="5"><a href="https://github.com/farmerx/sshtunnel" target="_blank"><i class="el-icon-message"></i>Star On Github</a></el-menu-item>
    </el-menu>
    <div class='content'>
        <div v-if="activeIndex === '1'">
            <div class="btn">
                <span class="wrapper">
                    <el-button type="success" icon="share"  @click="dialogAddRemoteVisible = true">添加映射</el-button>
                    <el-button type="success" icon="delete"  @click="deleteRemoteServer()">批量删除</el-button>
                </span>
            </div>
            <div class="table">
                <el-table ref="multipleTable" :data="list" border tooltip-effect="dark" style="width: 100%" @selection-change="handleSelectionChange">
                    <el-table-column type="selection" width="55"> </el-table-column>
                    <el-table-column label="UUID"> <template scope="scope">{{ scope.row['uuid'] }}</template></el-table-column>
                    <el-table-column prop="localaddr" label="本地监听端口" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="remoteaddr" label="映射端口地址" show-overflow-tooltip></el-table-column>
                    <el-table-column label="SSH端口地址" show-overflow-tooltip>
                      <template scope="scope">
                        <el-popover trigger="hover" placement="top">
                          <p>SSHUser: {{ scope.row.sshUser }}</p>
                          <p>SSHPwd: {{ scope.row.sshPwd }}</p>
                          <p>SSHPubkey: {{ scope.row.sshpubkey }}</p>
                          <div slot="reference" class="name-wrapper">
                            <el-tag>{{ scope.row.middleaddr }}</el-tag>
                          </div>
                        </el-popover>
                      </template>
                    </el-table-column>
                    <el-table-column label="状态"> 
                      <template scope="scope"> 
                        <div v-if="scope.row.ststus === true">
                             <el-button size="small" type="primary">运行中...</el-button>
                        </div>
                        <div v-if="scope.row.ststus ===false">
                             <el-button size="small" type="danger">已停止</el-button>
                        </div>
                       <!--<el-button size="small" type="primary">{{translate(scope.row.ststus)}}</el-button>-->
                     </template>
                    </el-table-column>
                    <el-table-column label="日志信息">
                      <template scope="scope">
                        <el-button
                          size="small"
                          type="info"
                          @click="showMsg(scope.$index, scope.row)"><i class="el-icon-warning"></i>日志详情...</el-button>
                      </template>
                    </el-table-column>
                    <el-table-column label="操作" >
                      <template scope="scope">
                        <span>
                        <el-button
                          size="small"
                          type="primary"
                          @click="handleEdit(scope.$index, scope.row)"><i class="el-icon-edit"></i></el-button>
                        </span>
                        <span>
                        <el-button
                          size="small"
                          type="danger"
                          @click="handleDelete(scope.$index, scope.row)"><i class="el-icon-delete"></i></el-button>
                        </span>
                        <span v-if="scope.row.ststus === true">
                              <el-button
                          size="small"
                          type="warning"
                          @click="handleOperate(scope.$index, scope.row)">停止</el-button>
                        </span>
                        <span v-if="scope.row.ststus ===false">
                             <el-button
                          size="small"
                          type="warning"
                          @click="handleOperate(scope.$index, scope.row)">启动</el-button>
                        </span>
                      </template>
                    </el-table-column>
                </el-table>
            </div>
        </div>
    </div>
    <el-dialog title="添加映射" :visible.sync="dialogAddRemoteVisible" size="tiny">
      <el-form ref="remoteServerform" :rules="rules" :label-position="labelPosition" :model="remoteServerform" label-width="120px">
        <el-form-item label="localAddr" prop="localAddr">
          <el-input v-model="remoteServerform.localAddr"></el-input>
        </el-form-item>
        <el-form-item label="remoteAddr" prop="remoteAddr">
          <el-input v-model="remoteServerform.remoteAddr"></el-input>
        </el-form-item>
        <el-form-item label="sshServerAddr"  prop="middleAddr">
          <el-input v-model="remoteServerform.middleAddr"></el-input>
        </el-form-item>
        <el-form-item label="ssh用户" prop="sshuser">
          <el-input v-model="remoteServerform.sshuser"></el-input>
        </el-form-item>
        <el-form-item label="ssh密码" prop="sshpwd">
          <el-input type="password" v-model="remoteServerform.sshpwd"></el-input>
        </el-form-item>
         <el-form-item label="ssh公钥文件路径" prop="sshpubkey">
          <el-input v-model="remoteServerform.sshpubkey"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary"  @click="addsubmit('remoteServerform')">立即添加</el-button>
          <el-button @click="resetForm('remoteServerform')">重置</el-button>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogAddRemoteVisible = false">取 消</el-button>
        <!--<el-button type="primary" @click="dialogAddRemoteVisible = false">确 定</el-button>-->
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
export default {
  name: 'dashboard',
  data () {
    var validatePass = (rule, value, callback) => {
      if (this.remoteServerform.sshpwd === '' && this.remoteServerform.sshpubkey === '') {
        return callback(new Error('ssh 密码和公钥不能同时为空.'))
      }
      callback()
    }
    return {
      labelPosition: 'right',
      remoteServerform: {
        url: '/addRemoteServer',
        localAddr: '',
        remoteAddr: '',
        middleAddr: '',
        sshuser: '',
        sshpwd: '',
        sshpubkey: '',
        uuid: ''
      },
      rules: {
        localAddr: [
          { required: true, message: '请输入本地监听端口', trigger: 'blur' }
        ],
        remoteAddr: [
          { required: true, message: '请输入远程服务映射端口', trigger: 'blur' }
        ],
        middleAddr: [
          { required: true, message: '请输入ssh服务host和port', trigger: 'blur' }
        ],
        sshuser: [
          { required: true, message: '请输入ssh登录的用户名', trigger: 'blur' }
        ],
        sshpwd: [
          { validator: validatePass, trigger: 'blur' }
        ],
        sshpubkey: [
          { validator: validatePass, trigger: 'blur' }
        ]
      },
      dialogAddRemoteVisible: false,
      activeIndex: '1',
      multipleSelection: []
    }
  },
  methods: {
    handleDelete (index, row) {
      alert('暂时不支持删除功能')
    },
    handleEdit (index, row) {
      this.remoteServerform = {
        url: '/addRemoteServer',
        localAddr: row.localaddr,
        remoteAddr: row.remoteaddr,
        middleAddr: row.middleaddr,
        sshuser: row.sshUser,
        sshpwd: row.sshPwd,
        sshpubkey: row.sshpubkey,
        uuid: row.uuid
      }
      this.dialogAddRemoteVisible = true
    },
    handleOperate (index, row) {
      if (row.ststus) {
        row['operate'] = 'stop'
      } else {
        row['operate'] = 'start'
      }
      row['url'] = '/operateRemoteServer'
      this.$store.dispatch('handleOperate', row)
    },
    addsubmit (formName) {
      this.$refs[formName].validate((valid) => {
        if (valid) {
          this.$store.dispatch('AddRemoteServer', this.remoteServerform)
          this.dialogAddRemoteVisible = false
        } else {
          console.log('error submit!!')
          return false
        }
      })
    },
    resetForm (formName) {
      this.$refs[formName].resetFields()
    },
    handleSelectionChange (val) {
      this.multipleSelection = val
    },
    addRemoteServer () {
      alert('xxxxx')
    },
    deleteRemoteServer () {
      alert('暂不支持批量删除')
    },
    translate (flag) {
      if (flag) {
        return '运行中'
      } else {
        return '停止'
      }
    },
    handleSelect (key, keyPath) {
      this.activeIndex = key
    }
  },
  computed: {
    ...mapGetters([
      'LoginError',
      'LoginFlag',
      'Action',
      'list',
      'ListErrorMsg',
      'AddServerErrorMsg'
    ])
  },
  watch: {
    Action () {
      if (this.LoginFlag === false) {
        this.$message.error(this.LoginError)
      } else {
        this.$router.push({ path: '/' })
      }
    },
    ListErrorMsg () {
      this.$message.error(this.ListErrorMsg)
    },
    AddServerErrorMsg () {
      this.$message.error(this.AddServerErrorMsg)
    }
  },
  created () {
    if (this.$cookie.get('SSHTunnelUserInfo') !== null) {
      this.$router.push({path: '/'})
    } else {
      var tmp = {'url': '/getremoteserverlist'}
      this.$store.dispatch('GetServerList', tmp)
    }
  }
}
</script>

<style scoped lang='less'>
    .dashboard{
        a{
            text-decoration:none; 
        }
        .el-menu {
            border-radius: 5px;
        }
        .content {
            // padding-top: 10px;
            .btn{
                margin-top: 5px;
                margin-bottom: 5px;
            }
        }
    }
</style>]

