const fs = require('fs');
const path = require('path');
const { chromium } = require('playwright');

// Функція для видалення частини URL після символу #
function cleanURL(url) {
  const hashIndex = url.indexOf('#');
  return hashIndex !== -1 ? url.substring(0, hashIndex) : url;
}

(async () => {
  // Шлях до файлу з об'єктами
  const jsonPath = path.join(__dirname, '../internal/objects/parsed_objects.json');
  const rawData = fs.readFileSync(jsonPath, 'utf-8');
  let objects;
  try {
    objects = JSON.parse(rawData);
  } catch (e) {
    console.error('❌ Помилка парсингу JSON:', e.message);
    process.exit(1);
  }

  // Перевірка наявності об'єктів
  if (!Array.isArray(objects) || objects.length === 0) {
    console.error('❌ У файлі немає обʼєктів');
    process.exit(1);
  }

  // Шлях до файлу з пошуковими URL
  const searchURLsPath = path.join(__dirname, '../internal/objects/search_URLs.json');
  if (!fs.existsSync(searchURLsPath)) {
    fs.writeFileSync(searchURLsPath, JSON.stringify([], null, 2));
  }
  const savedURLs = JSON.parse(fs.readFileSync(searchURLsPath, 'utf-8'));

  // Очищення дублікатів за URL
  const uniqueURLs = [];
  const seen = new Set();
  for (const entry of savedURLs) {
    if (!seen.has(entry.url)) {
      seen.add(entry.url);
      uniqueURLs.push(entry);
    }
  }

  if (uniqueURLs.length !== savedURLs.length) {
    fs.writeFileSync(searchURLsPath, JSON.stringify(uniqueURLs, null, 2), 'utf-8');
    console.log('🧹 Видалено дублікати зі search_URLs.json');
  }
  savedURLs.length = 0;
  savedURLs.push(...uniqueURLs);

  // Запуск браузера
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();

  // Обробка кожного об'єкту
  for (const obj of objects) {
    const { title, location, price, rooms, category, region, link } = obj;

    // Розділення назви на частини
    const commaIndex = title.indexOf(',');
    if (commaIndex === -1) {
      console.error('❌ Неможливо розділити title: відсутня кома');
      continue;
    }

    const part1 = title.slice(0, commaIndex).trim();
    const part2 = title.slice(commaIndex + 1).trim();
    let street = `${part1} (${location}), ${part2}`;
    console.log(`🔍 Вулиця для пошуку: ${street}`);

    // Отримання базового URL
    const categoryIndex = link.indexOf(`/${category}/`);
    if (categoryIndex === -1) {
      console.error(`❌ Не вдалося знайти category "${category}" у лінку: ${link}`);
      continue;
    }

    const baseURL = link.substring(0, categoryIndex + category.length + 2);
    console.log(`🌐 Перехід за URL: ${baseURL}`);
    await page.goto(baseURL, { timeout: 5000 });

    // Введення вулиці у пошук
    const inputSelector = 'input.nav_street_input';
    const inputLocator = page.locator(inputSelector);
    await inputLocator.waitFor();

    // Пошук вулиці з районом і з номером будинку
    let found = await trySearchStreet(page, street, inputSelector);

    // Пошук вулиці без району
    if (!found) {
      street = `${part1}, ${part2}`;
      console.log(`🔍 Вулиця для пошуку без району: ${street}`);
      found = await trySearchStreet(page, street, inputSelector);
    }

    // Пошук вулиці з районом, але без номеру будинку
    if (!found) {
      street = `${part1} (${location})`;
      console.log(`🔍 Вулиця для пошуку (район, без номеру): ${street}`);
      found = await trySearchStreet(page, street, inputSelector);
    }

    // Пошук лише вулиці без району і без номеру будинку
    if (!found) {
      street = `${part1}`;
      console.log(`🔍 Вулиця для пошуку (без району, без номеру): ${street}`);
      const searchPromise = trySearchStreet(page, street, inputSelector);
      const timeoutPromise = new Promise(resolve => setTimeout(() => resolve(false), 3000)); // Таймаут 3 секунди

      found = await Promise.race([searchPromise, timeoutPromise]);
      if (!found) {
        console.log(`⚠️ Вулиця "${street}" не знайдена за 3 секунди, переходжу до наступного етапу.`);
      }
    }

    // Пошук лише за районом
    if (!found) {
      street = `${location}`;
      console.log(`🔍 Пошук лише за районом: ${street}`);
      found = await trySearchStreet(page, street, inputSelector);
    }

    if (!found) continue;

    // Вибираємо вулицю з районом, якщо вона знайдена
    if (found) {
      if (street === location) {
        const districtOnly = location.replace(/р-н/g, 'район');
        await page.click(`div.nav_item_option:has-text("${districtOnly}")`);
      } else {
        await page.click(`div.nav_item_option[data-title="${street}"]`);
      }
      await page.waitForTimeout(500);
    }

    await page.click('div.nav_item_active[data-standart="Ціна"]');
    await page.waitForSelector('.currency-switcher', { timeout: 1000 });

    const cleanPriceStr = price.replace(/[^\d]/g, '');
    const priceTo = parseInt(cleanPriceStr, 10);
    const priceFrom = Math.round(priceTo * 0.85);

    let currencyCode = '2'; // Долар за замовчуванням
    if (price.includes('грн')) {
      currencyCode = '1';
    } else if (price.includes('$')) {
      currencyCode = '2';
    } else if (price.includes('€')) {
      currencyCode = '3';
    } else {
      console.warn(`⚠️ Невідомий символ валюти в рядку: "${price}", обираємо долар за замовчуванням`);
    }

    await page.click(`.currency-switcher__currency[data-currency="${currencyCode}"]`);
    await page.waitForTimeout(500);
    await page.fill('.filter-field__input.js_input_from', priceFrom.toString());
    await page.fill('.filter-field__input.js_input_to', priceTo.toString());
    await page.click('button.filter__apply-button:visible');
    await page.waitForTimeout(500);

    // Введення кількості кімнат, якщо це не будинок або земельна ділянка
    if (!['houses-sale', 'areas-sale'].includes(category)) {
      await page.click('div.nav_item_active[data-standart="Кімнат"]');
      await page.waitForSelector('.filter-room-count-options');

      const roomNumber = rooms.toLowerCase().includes('студія')
        ? 'Студія'
        : rooms.match(/\d+/)?.[0] || null;

      if (!roomNumber) {
        console.warn('⚠️ Неможливо визначити кількість кімнат');
      } else {
        const roomSelector =
          roomNumber === 'Студія'
            ? `.filter-room-count__option[data-title="Студія"]`
            : `.filter-room-count__option[data-title="${roomNumber}"]`;

        await page.locator(roomSelector).click({ force: true });
        await page.waitForTimeout(500);

        const applyButton = page.locator('button.filter__apply-button:visible').nth(2);
        await applyButton.click({ position: { x: 0, y: 0 }, force: true });
      }
    }

    await page.click('button.nav_search_btn');
    await page.waitForTimeout(1000);

    const finalURL = page.url();
    const cleanedFinalURL = cleanURL(finalURL); // Використання функції cleanURL
    console.log(`✅ URL для перевірки результатів: ${cleanedFinalURL}`);

    // Перевірка та збереження URL
    if (!savedURLs.some(item => item.url === cleanedFinalURL)) {
      savedURLs.push({ title, url: cleanedFinalURL });
      fs.writeFileSync(searchURLsPath, JSON.stringify(savedURLs, null, 2), 'utf-8');
      console.log('💾 URL збережено у search_URLs.json');
    } else {
      console.log('ℹ️ URL вже існує у search_URLs.json, пропускаємо збереження');
    }
  }

  await browser.close();
})();

// Функція для спроби пошуку вулиці
async function trySearchStreet(page, street, inputSelector) {
  const inputLocator = page.locator(inputSelector);
  await inputLocator.waitFor();
  await inputLocator.click();
  await inputLocator.fill(''); // Очищення поля перед кожним пошуком

  for (const char of street.slice(0, 4)) {
    await page.keyboard.type(char);
    await page.dispatchEvent(inputSelector, 'input');
    await page.dispatchEvent(inputSelector, 'keydown');
    await page.dispatchEvent(inputSelector, 'keyup');
    await page.waitForTimeout(500);
  }

  await page.waitForTimeout(1000);
  await page.waitForSelector('div.nav_item_option[data-title]', { timeout: 5000 });

  let items = await page.$$eval('div.nav_item_option[data-title]', options =>
    options.map(o => o.getAttribute('data-title'))
  );

  // Перевірка на наявність вулиці з урахуванням різних варіантів написання району
  const locationVariants = [
    street,
    street.replace(/р-н/g, 'район'),
    street.replace(/район/g, 'р-н')
  ];

  const foundVariant = items.some(item => locationVariants.some(variant => item.includes(variant)));
  if (foundVariant) {
    return true;
  } else {
    console.error(`\n❌ Вулиця не знайдена: "${street}"`);
    return false;
  }
}
