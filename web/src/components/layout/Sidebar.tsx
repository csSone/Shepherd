import { Link, useLocation } from 'react-router-dom';
import { useUIStore } from '@/stores/uiStore';
import { cn } from '@/lib/utils';
import {
  LayoutDashboard,
  Package,
  Download,
  MessageSquare,
  Network,
  ScrollText,
  Settings,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';

/**
 * 侧边栏导航项配置
 */
const navItems = [
  { path: '/', icon: LayoutDashboard, label: '仪表盘' },
  { path: '/models', icon: Package, label: '模型管理' },
  { path: '/downloads', icon: Download, label: '下载管理' },
  { path: '/chat', icon: MessageSquare, label: '聊天' },
  { path: '/cluster', icon: Network, label: '集群管理' },
  { path: '/logs', icon: ScrollText, label: '日志' },
  { path: '/settings', icon: Settings, label: '设置' },
];

/**
 * 侧边栏组件
 */
export function Sidebar() {
  const location = useLocation();
  const { sidebarOpen, toggleSidebar } = useUIStore();

  return (
    <aside
      className={cn(
        'flex flex-col border-r bg-background transition-all duration-300',
        sidebarOpen ? 'w-64' : 'w-16'
      )}
    >
      {/* Logo 区域 */}
      <div className="flex h-16 items-center justify-between border-b px-4">
        {sidebarOpen && (
          <Link to="/" className="flex items-center gap-2 font-semibold">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              🐏
            </div>
            <span>Shepherd</span>
          </Link>
        )}
        <button
          onClick={toggleSidebar}
          className="ml-auto rounded-lg p-2 hover:bg-accent"
          aria-label={sidebarOpen ? '收起侧边栏' : '展开侧边栏'}
        >
          {sidebarOpen ? <ChevronLeft size={18} /> : <ChevronRight size={18} />}
        </button>
      </div>

      {/* 导航菜单 */}
      <nav className="flex-1 overflow-y-auto py-4">
        <ul className="space-y-1 px-2">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = location.pathname === item.path;

            return (
              <li key={item.path}>
                <Link
                  to={item.path}
                  className={cn(
                    'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                  )}
                >
                  <Icon size={20} />
                  {sidebarOpen && <span>{item.label}</span>}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      {/* 底部信息 */}
      <div className="border-t p-4">
        {sidebarOpen && (
          <div className="text-xs text-muted-foreground">
            <div>Shepherd v0.1.0-alpha</div>
            <div className="mt-1">© 2026 Shepherd Project</div>
          </div>
        )}
      </div>
    </aside>
  );
}
