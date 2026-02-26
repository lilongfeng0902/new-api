// 在浏览器控制台执行此脚本验证按钮是否显示

console.log('========== 验证短信登录按钮 ==========');

// 1. 检查 localStorage
const status = JSON.parse(localStorage.getItem('status') || '{}');
console.log('✓ SMS Login Enabled:', status.sms_login_enabled);
console.log('✓ SMS Bind Enabled:', status.sms_bind_enabled);

// 2. 检查DOM中是否有按钮
setTimeout(() => {
  const buttons = Array.from(document.querySelectorAll('button'));
  const smsButton = buttons.find(btn => btn.textContent.includes('手机号'));

  if (smsButton) {
    console.log('✅ 找到短信登录按钮!', smsButton);
    smsButton.style.border = '3px solid red'; // 高亮显示
  } else {
    console.log('❌ 未找到短信登录按钮');
    console.log('当前页面的所有按钮:', buttons.map(b => b.textContent.trim()).filter(t => t));
  }
}, 1000);

// 3. 检查是否在登录页面
if (window.location.pathname !== '/login') {
  console.warn('⚠️  当前不在登录页面，请访问: http://localhost:3000/login');
}
