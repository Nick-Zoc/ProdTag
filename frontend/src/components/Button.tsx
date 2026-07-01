import type {ButtonHTMLAttributes} from 'react';
import {classNames} from '../utils/classNames';

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'success';

const variantClass: Record<ButtonVariant, string> = {
  primary: 'button-primary',
  secondary: 'button-secondary',
  ghost: 'button-ghost',
  danger: 'button-danger',
  success: 'button-success',
};

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
};

export function Button({children, className, variant = 'primary', type = 'button', ...props}: ButtonProps) {
  return (
    <button type={type} className={classNames('button', variantClass[variant], className)} {...props}>
      {children}
    </button>
  );
}
