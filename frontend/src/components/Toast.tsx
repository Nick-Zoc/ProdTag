import {Badge} from './Badge';

export type ToastTone = 'neutral' | 'green' | 'amber' | 'rose';

export type ToastState = {
  message: string;
  tone: ToastTone;
};

export function Toast({toast}: {toast: ToastState}) {
  return (
    <div className="toast">
      <span className="text-sm text-neutral-700">{toast.message}</span>
      <Badge tone={toast.tone}>{toastLabel(toast.tone)}</Badge>
    </div>
  );
}

function toastLabel(tone: ToastTone) {
  if (tone === 'green') {
    return 'Done';
  }
  if (tone === 'amber') {
    return 'Working';
  }
  if (tone === 'rose') {
    return 'Error';
  }
  return 'Info';
}
