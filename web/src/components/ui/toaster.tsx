import { useToastStore } from '@/stores/toast';
import { Toast } from './toast';

export function Toaster() {
  const { toasts, removeToast } = useToastStore();

  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none">
      {toasts.map((toast) => (
        <div key={toast.id} className="pointer-events-auto">
          <Toast {...toast} onClose={() => removeToast(toast.id)} />
        </div>
      ))}
    </div>
  );
}
