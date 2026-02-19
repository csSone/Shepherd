import { Bell, Search, User } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useUIStore } from '@/stores/uiStore';

/**
 * 顶部栏组件
 */
export function Header() {
  const { searchQuery, setSearchQuery } = useUIStore();

  return (
    <header className="flex h-16 items-center justify-between border-b bg-background px-6">
      {/* 搜索框 */}
      <div className="flex items-center gap-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="search"
            placeholder="搜索..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-10 w-64 rounded-lg border bg-background pl-10 pr-4 text-sm outline-none focus:ring-2 focus:ring-ring"
          />
        </div>
      </div>

      {/* 右侧操作 */}
      <div className="flex items-center gap-2">
        {/* 通知按钮 */}
        <Button variant="ghost" size="icon" aria-label="通知">
          <Bell size={20} />
        </Button>

        {/* 用户菜单 */}
        <Button variant="ghost" size="icon" aria-label="用户">
          <User size={20} />
        </Button>
      </div>
    </header>
  );
}
