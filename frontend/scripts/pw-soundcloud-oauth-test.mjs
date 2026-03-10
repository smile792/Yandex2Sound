import { chromium } from 'playwright';

const BASE = 'http://localhost:5173';

const result = {
  baseUrl: BASE,
  startedAt: new Date().toISOString(),
  steps: [],
  finalUrl: '',
  success: false,
  error: ''
};

const browser = await chromium.launch({ headless: true });
const context = await browser.newContext();
const page = await context.newPage();

try {
  result.steps.push('open_home');
  await page.goto(BASE, { waitUntil: 'domcontentloaded', timeout: 30000 });

  // Click SoundCloud OAuth button (second connect button)
  const soundcloudButton = page.getByRole('button', { name: 'Connect via OAuth' }).nth(1);
  await soundcloudButton.waitFor({ timeout: 15000 });
  result.steps.push('click_soundcloud_connect');
  await soundcloudButton.click();

  await page.waitForTimeout(4000);
  const urlAfterClick = page.url();
  result.steps.push(`url_after_click:${urlAfterClick}`);

  // Wait up to 25s for either callback or error page
  const deadline = Date.now() + 25000;
  while (Date.now() < deadline) {
    const u = page.url();
    if (u.includes('/api/soundcloud/auth/callback')) {
      result.steps.push('reached_callback');
      break;
    }
    if (u.includes('soundcloud.com') || u.includes('secure.soundcloud.com') || u.includes('api-auth.soundcloud.com')) {
      // still in oauth flow
    }
    await page.waitForTimeout(500);
  }

  result.finalUrl = page.url();
  const bodyText = (await page.locator('body').innerText().catch(() => '')).slice(0, 500);
  result.steps.push(`body_preview:${bodyText.replace(/\s+/g, ' ').trim()}`);

  // Consider success if redirected back to frontend with sc=connected
  if (result.finalUrl.includes('sc=connected')) {
    result.success = true;
    result.steps.push('soundcloud_connected');
  }
} catch (e) {
  result.error = e instanceof Error ? `${e.name}: ${e.message}` : String(e);
} finally {
  await context.close();
  await browser.close();
}

console.log(JSON.stringify(result, null, 2));
