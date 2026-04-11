import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuthStore } from '../../store/authStore';
import { Button } from '../../components/common/Button';
import { Input } from '../../components/common/Input';
import { toast } from '../../components/common/Toast';

const loginSchema = z.object({
  username: z.string().min(1, '请输入账号'),
  password: z.string().min(1, '请输入密码'),
});

export default function Login() {
  const navigate = useNavigate();
  const { login, loading } = useAuthStore();
  const [error, setError] = useState('');

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data) => {
    setError('');
    const success = await login(data);
    if (success) {
      toast.success('登录成功');
      navigate('/');
    } else {
      setError(useAuthStore.getState().error || '登录失败');
    }
  };

  return (
    <div className="max-w-md mx-auto">
      <div className="bg-white rounded-xl shadow-sm p-8">
        <h2 className="text-2xl font-bold text-center text-gray-900 mb-8">用户登录</h2>

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

          <Button
            type="submit"
            loading={loading}
            className="w-full"
            size="large"
          >
            登录
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-gray-600">
          还没有账号？{' '}
          <Link to="/register" className="text-blue-600 hover:text-blue-700">
            立即注册
          </Link>
        </p>
      </div>
    </div>
  );
}