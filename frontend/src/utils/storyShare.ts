import { gradients } from './gradients';
import { _ } from 'svelte-i18n';
import { get } from 'svelte/store';

export interface HabitStoryImageOptions {
  width?: number;
  height?: number;
  habitTitle: string;
  streakDays: number;
  wantToBecomeText?: string;
  wantLabelText?: string; // e.g. «Кем хочу стать»
  gradientCss?: string; // e.g. 'linear-gradient(135deg, #ff1b6b 0%, #45caff 100%)'
}

export function getGradientForHabitId(habitId: string): string {
  const index = simpleHash(habitId) % gradients.length;
  return gradients[index];
}

function simpleHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash;
  }
  return Math.abs(hash);
}

function parseGradientColors(gradientCss?: string): [string, string] {
  if (!gradientCss) {
    return ['#ff1b6b', '#45caff'];
  }
  const matches = gradientCss.match(/#[0-9a-fA-F]{3,8}/g) || [];
  if (matches.length >= 2) {
    return [matches[0], matches[1]] as [string, string];
  }
  return ['#ff1b6b', '#45caff'];
}

function drawRoundedRectPath(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  radius: number
) {
  const r = Math.min(radius, width / 2, height / 2);
  ctx.beginPath();
  ctx.moveTo(x + r, y);
  ctx.lineTo(x + width - r, y);
  ctx.quadraticCurveTo(x + width, y, x + width, y + r);
  ctx.lineTo(x + width, y + height - r);
  ctx.quadraticCurveTo(x + width, y + height, x + width - r, y + height);
  ctx.lineTo(x + r, y + height);
  ctx.quadraticCurveTo(x, y + height, x, y + height - r);
  ctx.lineTo(x, y + r);
  ctx.quadraticCurveTo(x, y, x + r, y);
  ctx.closePath();
}

function wrapText(
  ctx: CanvasRenderingContext2D,
  text: string,
  x: number,
  y: number,
  maxWidth: number,
  lineHeight: number,
  maxLines: number
): number {
  const words = text.split(/\s+/);
  let line = '';
  let lineCount = 0;
  for (let n = 0; n < words.length; n++) {
    const testLine = line ? `${line} ${words[n]}` : words[n];
    const metrics = ctx.measureText(testLine);
    if (metrics.width > maxWidth && n > 0) {
      ctx.fillText(line, x, y);
      line = words[n];
      y += lineHeight;
      lineCount += 1;
      if (lineCount >= maxLines - 1) {
        // last line with ellipsis
        let rest = line;
        while (ctx.measureText(`${rest}…`).width > maxWidth && rest.length > 0) {
          rest = rest.slice(0, -1);
        }
        ctx.fillText(`${rest}…`, x, y);
        return y + lineHeight;
      }
    } else {
      line = testLine;
    }
  }
  ctx.fillText(line, x, y);
  return y + lineHeight;
}

const SQUIRCLE_URL = new URL('../assets/squircley.svg', import.meta.url).href;

async function ensureFontsLoaded(): Promise<void> {
  try {
    // Wait for document fonts if supported
    if ((document as any).fonts?.ready) {
      await (document as any).fonts.ready;
    }
    // Additionally try to load specific faces
    if ((document as any).fonts?.load) {
      await Promise.all([
        (document as any).fonts.load('bold 64px Manrope'),
        (document as any).fonts.load('600 56px Manrope'),
        (document as any).fonts.load('500 36px Manrope'),
      ]);
    }
  } catch {
    // ignore
  }
}

function loadImage(url: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.crossOrigin = 'anonymous';
    img.onload = () => resolve(img);
    img.onerror = reject;
    img.src = url;
  });
}

async function drawMaskedSquircleCard(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  w: number,
  h: number,
  renderContent: (c: CanvasRenderingContext2D, width: number, height: number) => void,
): Promise<void> {
  const off = document.createElement('canvas');
  off.width = Math.round(w);
  off.height = Math.round(h);
  const octx = off.getContext('2d');
  if (!octx) return;

  // Draw content
  renderContent(octx, off.width, off.height);

  // Clip with squircle mask
  const mask = await loadImage(SQUIRCLE_URL);
  octx.globalCompositeOperation = 'destination-in';
  // Fit mask onto offscreen canvas preserving aspect ratio
  octx.drawImage(mask, 0, 0, off.width, off.height);
  octx.globalCompositeOperation = 'source-over';

  // Draw onto main canvas
  ctx.drawImage(off, x, y, w, h);
}

export async function generateHabitStoryImage(options: HabitStoryImageOptions): Promise<string> {
  await ensureFontsLoaded();
  const width = options.width ?? 1080;
  const height = options.height ?? 1920;
  const [c1, c2] = parseGradientColors(options.gradientCss);

  const canvas = document.createElement('canvas');
  canvas.width = width;
  canvas.height = height;
  const ctx = canvas.getContext('2d');
  if (!ctx) throw new Error('Canvas is not supported');

  // Background gradient (approx 135deg)
  const grad = ctx.createLinearGradient(0, height, width, 0);
  grad.addColorStop(0, c1);
  grad.addColorStop(1, c2);
  ctx.fillStyle = grad;
  ctx.fillRect(0, 0, width, height);

  // Invitation text (centered)
  const offsetY = 60; // общий сдвиг всего контента вниз
  ctx.fillStyle = '#fafafa';
  ctx.font = '700 56px Manrope, Inter, system-ui, -apple-system, sans-serif';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'top';
  const t = get(_);
  const invite = t('habits_story.story_invite') as string;
  const inviteY = 180 + offsetY;
  // Печатаем весь текст без усечения: увеличиваем maxLines
  wrapTextCentered(ctx, invite, width / 2, inviteY, Math.min(980, width - 120), 64, 10);

  // Habit card (squircle) under text
  const cardWidth = Math.min(860, width - 200);
  const cardHeight = Math.round(cardWidth * 0.88); // немного сплющенный по высоте
  const cardX = (width - cardWidth) / 2;
  const cardY = inviteY + 180;

  await drawMaskedSquircleCard(ctx, cardX, cardY, cardWidth, cardHeight, (c, w, h) => {
    // white card background
    c.fillStyle = 'rgba(255,255,255,0.96)';
    c.fillRect(0, 0, w, h);

    // content center
    const padX = 32; // как в HabitCard
    let cy = Math.round(h * 0.40); // ещё выше заголовок
    c.textAlign = 'center';
    c.textBaseline = 'alphabetic';
    c.fillStyle = '#111';
    c.font = '700 64px Manrope, Inter, system-ui, -apple-system, sans-serif';
    cy = wrapTextCentered(c, options.habitTitle, w / 2, cy, w - padX * 2, 72, 2);

    // want to become block
    if (options.wantToBecomeText) {
      cy += 18; // чуть опустил label
      c.fillStyle = 'rgba(17,17,17,0.6)';
      c.font = '500 28px Manrope, Inter, system-ui, -apple-system, sans-serif';
      c.fillText(options.wantLabelText || 'Кем хочу стать', w / 2, cy);
      cy += 126; // ещё ниже текст want
      c.fillStyle = '#111';
      // тот же размер, что и у названия
      c.font = '700 64px Manrope, Inter, system-ui, -apple-system, sans-serif';
      wrapTextCentered(c, options.wantToBecomeText, w / 2, cy, w - padX * 2, 72, 2);
    }
  });

  // streak ниже карточки
  const streakText = `${options.streakDays} ${t('habits_story.days_in_row') as string}`;
  ctx.font = '700 48px Manrope, Inter, system-ui, -apple-system, sans-serif';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  const m = ctx.measureText(streakText);
  const pillPadX = 28;
  const pillPadY = 18;
  const pillW = Math.ceil(m.width + pillPadX * 2);
  const pillH = 72;
  const pillX = Math.round((width - pillW) / 2);
  const pillY = Math.round(cardY + cardHeight + 44);
  drawRoundedRectPath(ctx as any, pillX, pillY, pillW, pillH, 24);
  ctx.fillStyle = 'rgba(255,255,255,0.95)';
  ctx.fill();
  ctx.fillStyle = '#111827';
  ctx.fillText(streakText, pillX + pillW / 2, pillY + pillH / 2);

  // Кривая стрелка от правого низа карточки к зоне ссылки (низ экрана по центру)
  const sx = cardX + cardWidth - 24;
  const sy = cardY + cardHeight - 24;
  const ex = Math.round(width * 0.72); // конец стрелки существенно правее центра
  const ey = Math.round(height - 520 + offsetY); // двигаем конец стрелки вместе с остальным вниз
  const c1x = sx + 200;
  const c1y = sy + 60;
  const c2x = Math.min(width - 60, ex + 240); // вынос вправо, с ограничением кромки
  const c2y = ey - 60;
  drawCurvedArrow(ctx, sx, sy, ex, ey, c1x, c1y, c2x, c2y);

  return canvas.toDataURL('image/png', 0.95);
}

export interface UploadResult {
  url: string;
  deleteUrl?: string;
}

function wrapTextCentered(
  ctx: CanvasRenderingContext2D,
  text: string,
  centerX: number,
  y: number,
  maxWidth: number,
  lineHeight: number,
  maxLines: number
): number {
  const words = text.split(/\s+/);
  let line = '';
  let lineCount = 0;
  for (let n = 0; n < words.length; n++) {
    const testLine = line ? `${line} ${words[n]}` : words[n];
    const metrics = ctx.measureText(testLine);
    if (metrics.width > maxWidth && n > 0) {
      ctx.fillText(line, centerX, y);
      line = words[n];
      y += lineHeight;
      lineCount += 1;
      if (lineCount >= maxLines - 1) {
        let rest = line;
        while (ctx.measureText(`${rest}…`).width > maxWidth && rest.length > 0) {
          rest = rest.slice(0, -1);
        }
        ctx.fillText(`${rest}…`, centerX, y);
        return y + lineHeight;
      }
    } else {
      line = testLine;
    }
  }
  ctx.fillText(line, centerX, y);
  return y + lineHeight;
}

export async function uploadBase64ToImgbb(
  base64DataUrl: string,
  apiKey: string,
  opts?: { name?: string; expirationSeconds?: number }
): Promise<UploadResult> {
  const base64 = base64DataUrl.replace(/^data:image\/\w+;base64,/, '');
  const form = new FormData();
  form.append('image', base64);
  if (opts?.name) form.append('name', opts.name);

  const params = new URLSearchParams({ key: apiKey });
  if (opts?.expirationSeconds) {
    params.set('expiration', String(opts.expirationSeconds));
  }

  const resp = await fetch(`https://api.imgbb.com/1/upload?${params.toString()}`, {
    method: 'POST',
    body: form,
  });

  if (!resp.ok) {
    const text = await resp.text().catch(() => '');
    throw new Error(`imgbb upload failed: ${resp.status} ${text}`);
  }
  const json = await resp.json();
  if (!json?.success) {
    throw new Error('imgbb upload failed: response success=false');
  }
  return {
    url: json.data?.url || json.data?.display_url,
    deleteUrl: json.data?.delete_url,
  };
}


function drawCurvedArrow(
  ctx: CanvasRenderingContext2D,
  sx: number,
  sy: number,
  ex: number,
  ey: number,
  c1x: number,
  c1y: number,
  c2x: number,
  c2y: number
) {
  ctx.save();
  ctx.strokeStyle = 'rgba(255,255,255,0.95)';
  ctx.lineWidth = 8;
  ctx.lineCap = 'round';
  ctx.beginPath();
  ctx.moveTo(sx, sy);
  ctx.bezierCurveTo(c1x, c1y, c2x, c2y, ex, ey);
  ctx.stroke();

  // Arrowhead
  // Get direction at end of bezier using derivative approximation
  const t = 0.97;
  const dx = bezierDerivative(sx, c1x, c2x, ex, t);
  const dy = bezierDerivative(sy, c1y, c2y, ey, t);
  const angle = Math.atan2(dy, dx);
  const size = 22;
  ctx.beginPath();
  ctx.moveTo(ex, ey);
  ctx.lineTo(ex - size * Math.cos(angle - Math.PI / 7), ey - size * Math.sin(angle - Math.PI / 7));
  ctx.moveTo(ex, ey);
  ctx.lineTo(ex - size * Math.cos(angle + Math.PI / 7), ey - size * Math.sin(angle + Math.PI / 7));
  ctx.stroke();
  ctx.restore();
}

function bezierDerivative(p0: number, p1: number, p2: number, p3: number, t: number): number {
  // derivative of cubic Bezier: 3*(1-t)^2*(p1-p0) + 6*(1-t)*t*(p2-p1) + 3*t^2*(p3-p2)
  const mt = 1 - t;
  return 3 * mt * mt * (p1 - p0) + 6 * mt * t * (p2 - p1) + 3 * t * t * (p3 - p2);
}


