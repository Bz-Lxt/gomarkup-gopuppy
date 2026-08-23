import { test, expect } from '@playwright/test';
import { request } from '@playwright/test';

const dad = { email: 'dad@gopuppy.test', password: 'Puppy123!' };
const mom = { email: 'mom@gopuppy.test', password: 'Puppy123!' };

test.describe.configure({ mode: 'serial' });

async function login(page, acc: { email: string; password: string }) {
  await page.goto('/login');
  await page.getByLabel('邮箱').fill(acc.email);
  await page.getByLabel('密码').fill(acc.password);
  await page.getByRole('button', { name: /进入林家小院/ }).click();
  await expect(page.getByRole('heading', { name: /林家小院/ })).toBeVisible({ timeout: 20000 });
}

test('① 登录后看到家庭宠物与精确到天的年龄', async ({ page }) => {
  await login(page, dad);
  await expect(page.getByRole('heading', { name: '奶油' })).toBeVisible();
  await expect(page.getByText(/共 \d+ 天/).first()).toBeVisible();
});

test('② 今日喂食打卡', async ({ page }) => {
  await login(page, dad);
  await page.locator('article').filter({ hasText: '奶油' }).getByRole('button', { name: /午/ }).first().click();
  await expect(page.getByText('林爸爸').first()).toBeVisible({ timeout: 10000 });
});

test('③ 双端 WebSocket：妈妈端看到爸爸打卡', async ({ browser }) => {
  const dadCtx = await browser.newContext();
  const momCtx = await browser.newContext();
  const dadPage = await dadCtx.newPage();
  const momPage = await momCtx.newPage();
  await login(dadPage, dad);
  await login(momPage, mom);
  await dadPage.locator('article').filter({ hasText: '奶油' }).getByRole('button', { name: /晚/ }).first().click();
  await expect(momPage.locator('article').filter({ hasText: '奶油' }).getByText('林爸爸').first()).toBeVisible({ timeout: 8000 });
  await dadCtx.close();
  await momCtx.close();
});

test('④ 补录疫苗事件后时间轴可见', async ({ page }) => {
  await login(page, dad);
  await page.goto('/health');
  await expect(page.getByRole('heading', { name: /时间轴/ })).toBeVisible({ timeout: 15000 });
  await expect(page.getByText(/绝育|狂犬/).first()).toBeVisible();
  await page.getByRole('button', { name: '补录事件' }).click();
  await page.getByLabel('标题').fill('E2E加强针');
  await page.getByRole('button', { name: '保存' }).click();
  await expect(page.getByText('E2E加强针')).toBeVisible({ timeout: 10000 });
});

test('⑤ 相册页可打开且图表页非空', async ({ page }) => {
  await login(page, dad);
  await page.goto('/finance');
  await expect(page.getByRole('heading', { name: /账本与体重/ })).toBeVisible({ timeout: 15000 });
  await expect(page.locator('canvas').first()).toBeVisible();
  await page.goto('/album');
  await expect(page.getByRole('heading', { name: /云端相册/ })).toBeVisible();
});
