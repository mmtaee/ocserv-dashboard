import type { MessageSchema } from '@/locales/en'

export default {
  common: {
    appName: 'Ocserv 控制台', language: '語言', loading: '載入中', close: '關閉', more: '更多',
    breadcrumb: '麵包屑導覽', sidebar: '側邊欄', toggleSidebar: '切換側邊欄', mobileSidebarDescription: '顯示行動版側邊欄。',
  },
  auth: {
    title: '歡迎回來', subtitle: '登入以管理您的 Ocserv 基礎設施。', failure: '登入失敗',
    username: '使用者名稱', usernamePlaceholder: 'admin', password: '密碼', verification: '驗證',
    captchaPrompt: '請先完成驗證碼再登入。', rememberMe: '記住我', submit: '登入',
    promoTitle: '安全的基礎設施管理', promoDescription: '在單一控制台中監控工作階段、使用者、代理程式及系統健康狀態。',
  },
  setup: {
    title: '初始化控制台', description: '設定連線設定檔、保留政策及選用的驗證碼保護。',
    failure: '初始化失敗', serverAddress: '伺服器位址', serverAddressPlaceholder: 'vpn.example.com',
    serverPort: '伺服器連接埠', connectionName: '連線名稱', defaultConnectionName: 'Ocserv VPN',
    retention: '非活躍使用者保留期', retentionHelp: '刪除非活躍使用者前保留的天數。',
    autoDelete: '自動刪除非活躍使用者', autoDeleteHelp: '自動套用保留期限。',
    captchaSiteKey: '驗證碼網站金鑰', captchaSecretKey: '驗證碼密鑰', submit: '儲存並繼續',
    promoTitle: '為您的私人網路開啟全新起點。', promoDescription: '這些設定會直接傳送至產生的控制台系統 API。',
  },
  dashboard: {
    administration: '管理', overview: '基礎設施概覽', title: '控制台',
    connectedUsers: '已連線使用者', activeSessions: '活躍工作階段', managedServers: '受管伺服器',
    activityTitle: '服務活動', activityDescription: '服務回報活動後，控制台指標將顯示於此。',
  },
  unavailable: {
    title: '伺服器無法使用', description: '控制台無法載入系統初始化設定。', connectionFailed: '連線失敗', retry: '重試',
  },
  footer: {
    developedBy: '由 {developer} 開發', with: '用', love: '愛', reportIssue: '回報問題',
  },
  navigation: {
    search: '搜尋', searchPlaceholder: '搜尋文件...', documentation: '文件',
    gettingStarted: '開始使用', installation: '安裝', projectStructure: '專案結構', buildingApp: '建置應用程式',
    routing: '路由', dataFetching: '資料擷取', rendering: '呈現', caching: '快取', styling: '樣式',
    optimizing: '最佳化', configuring: '設定', testing: '測試', authentication: '驗證', deploying: '部署',
    upgrading: '升級', examples: '範例', apiReference: 'API 參考', components: '元件', fileConventions: '檔案慣例',
    functions: '函式', configOptions: '設定選項', cli: 'CLI', edgeRuntime: '邊緣執行環境', architecture: '架構',
    accessibility: '無障礙', fastRefresh: '快速重新整理', compiler: '編譯器', supportedBrowsers: '支援的瀏覽器', turbopack: 'Turbopack',
  },
} satisfies MessageSchema
