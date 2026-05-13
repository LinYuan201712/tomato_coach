# tomatofrontend

```
project-root/
├── public/                        # 公共静态资源
├── src/                           # 源代码目录
│   ├── assets/                    # 静态资源
│   │   └── logo.png               # 项目logo图片
│   ├── components/                # 公共组件目录
│   │   ├── Login/                 # 登录相关组件
│   │   │   ├── AuthForm.vue       # 认证表单容器组件
│   │   │   └── FormInput.vue      # 表单输入框组件
│   │   ├── PersonalCenter/
│   │   │   ├── UserProfile.vue
│   │   │   ├── UserSetting.vue
│   │   │   └── Userstatistics.vue
│   ├── router/                    # 路由配置
│   │   └── index.js               # 路由主文件
│   ├── views/                     # 页面视图组件
│   │   ├── Login.vue              # 登录页面
│   │   ├── Register.vue           # 注册页面
│   │   ├──CreateRoomView.vue
│   │   ├──HomeView.vue
│   │   ├──JoinRoomView.vue
│   │   ├──PersonalCenter.vue
│   │   ├──StudyRoomView.vue
│   │   └── TaskManagementView.vue
│   ├── App.vue                    # 应用根组件
│   └── main.js                    # 应用入口文件
├── .gitignore                     # Git忽略配置
├── babel.config.js                # Babel配置
├── jsconfig.json                  # JavaScript配置
├── package-lock.json              # 依赖版本锁定文件
├── package.json                   # 项目配置和依赖
├── README.md                      # 项目说明文档
└── vue.config.js                  # Vue CLI配置
```

## Project setup
```
npm install
```

### Compiles and hot-reloads for development
```
npm run serve
```

### Compiles and minifies for production
```
npm run build
```

### Lints and fixes files
```
npm run lint
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).
