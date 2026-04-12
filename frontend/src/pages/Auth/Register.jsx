import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuthStore } from '../../store/authStore';
import { Button } from '../../components/common/Button';
import { Input } from '../../components/common/Input';
import { toast } from '../../components/common/Toast';

const registerSchema = z.object({
  username: z.string().min(1, '请输入账号'),
  password: z.string().min(6, '密码至少6位'),
  confirmPassword: z.string().min(1, '请确认密码'),
}).refine((data) => data.password === data.confirmPassword, {
  message: '两次输入的密码不一致',
  path: ['confirmPassword'],
});

export default function Register() {
  const navigate = useNavigate();
  const { register: registerUser, loading } = useAuthStore();
  const [error, setError] = useState('');

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data) => {
    setError('');
    const result = await registerUser({
      username: data.username,
      password: data.password,
    });
    if (result) {
      toast.success('注册成功，请登录');
      navigate('/login');
    } else {
      setError(useAuthStore.getState().error || '注册失败');
    }
  };

  return (
    <div className="max-w-md mx-auto">
      <div className="bg-white rounded-xl shadow-sm p-8">
        <h2 className="text-2xl font-bold text-center text-gray-900 mb-8">用户注册</h2>

        {error && (
          <div className="mb-4 p-3 bg-red-50 text-red-600 rounded-lg text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <Input
            label="账号"
            placeholder="请输入账号"
            error={errors.username?.message}
            {...register('username')}
          />

          <Input
            label="密码"
            type="password"
            placeholder="请输入密码"
            error={errors.password?.message}
            {...register('password')}
          />

          <Input
            label="确认密码"
            type="password"
            placeholder="请再次输入密码"
            error={errors.confirmPassword?.message}
            {...register('confirmPassword')}
          />

          <Button
            type="submit"
            loading={loading}
            className="w-full"
            size="large"
          >
            注册
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-gray-600">
          已有账号？{' '}
          <Link to="/login" className="text-blue-600 hover:text-blue-700">
            立即登录
          </Link>
        </p>
      </div>
    </div>
  );
}