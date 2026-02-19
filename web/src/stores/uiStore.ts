import { create } from 'zustand';
import { persist } from 'zustand/middleware';

/**
 * UI 状态接口
 */
interface UIState {
  // 侧边栏
  sidebarOpen: boolean;
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;

  // 主题
  theme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark') => void;

  // 当前视图
  currentView: string;
  setCurrentView: (view: string) => void;

  // 模态框
  activeModal: string | null;
  openModal: (modal: string) => void;
  closeModal: () => void;

  // 搜索
  searchQuery: string;
  setSearchQuery: (query: string) => void;

  // 过滤器
  modelStatusFilter: string;
  setModelStatusFilter: (status: string) => void;
  showFavouritesOnly: boolean;
  setShowFavouritesOnly: (show: boolean) => void;
}

/**
 * UI 状态 Store
 */
export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      // 侧边栏
      sidebarOpen: true,
      toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
      setSidebarOpen: (open) => set({ sidebarOpen: open }),

      // 主题
      theme: 'light',
      setTheme: (theme) => set({ theme }),

      // 当前视图
      currentView: 'dashboard',
      setCurrentView: (currentView) => set({ currentView }),

      // 模态框
      activeModal: null,
      openModal: (activeModal) => set({ activeModal }),
      closeModal: () => set({ activeModal: null }),

      // 搜索
      searchQuery: '',
      setSearchQuery: (searchQuery) => set({ searchQuery }),

      // 过滤器
      modelStatusFilter: 'all',
      setModelStatusFilter: (modelStatusFilter) => set({ modelStatusFilter }),
      showFavouritesOnly: false,
      setShowFavouritesOnly: (showFavouritesOnly) => set({ showFavouritesOnly }),
    }),
    {
      name: 'shepherd-ui-storage',
      partialize: (state) => ({
        theme: state.theme,
        sidebarOpen: state.sidebarOpen,
      }),
    }
  )
);
