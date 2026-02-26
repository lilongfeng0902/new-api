// ========================================
// 清除前端缓存并重新加载状态
// ========================================
// 使用方法：在浏览器控制台（F12 -> Console）中粘贴并执行此脚本

// 方法1：清除status缓存（推荐）
console.log('=== 方法1：清除status缓存 ===');
localStorage.removeItem('status');
console.log('✅ status缓存已清除');
console.log('📌 现在刷新页面: location.reload()');

// 如果需要，可以取消注释下面的代码自动刷新
// location.reload();

// ========================================

// 方法2：重新从API获取status（高级）
console.log('\n=== 方法2：重新加载status ===');
fetch('/api/status')
  .then(res => res.json())
  .then(result => {
    if (result.success) {
      const newStatus = result.data;

      // 保存到localStorage
      localStorage.setItem('status', JSON.stringify(newStatus));

      console.log('✅ Status已更新:', {
        sms_login_enabled: newStatus.sms_login_enabled,
        sms_bind_enabled: newStatus.sms_bind_enabled,
        sms_reset_enabled: newStatus.sms_reset_enabled,
        register_enabled: newStatus.register_enabled
      });

      console.log('📌 现在刷新页面查看效果: location.reload()');

      // 自动刷新（可选）
      // location.reload();
    } else {
      console.error('❌ 获取status失败:', result.message);
    }
  })
  .catch(err => {
    console.error('❌ 请求失败:', err);
  });

// ========================================

// 方法3：查看当前status（调试用）
console.log('\n=== 当前localStorage中的status ===');
const currentStatus = localStorage.getItem('status');
if (currentStatus) {
  const parsedStatus = JSON.parse(currentStatus);
  console.log('SMS相关配置:', {
    sms_login_enabled: parsedStatus.sms_login_enabled || '❌ 不存在',
    sms_bind_enabled: parsedStatus.sms_bind_enabled || '❌ 不存在',
    sms_reset_enabled: parsedStatus.sms_reset_enabled || '❌ 不存在',
    register_enabled: parsedStatus.register_enabled || '❌ 不存在'
  });
} else {
  console.log('❌ localStorage中没有status数据');
}
