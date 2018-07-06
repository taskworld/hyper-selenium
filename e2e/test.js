const { Builder, By, Key, until } = require('selenium-webdriver')
;(async function example() {
  console.log('>>> Initializing driver...')
  const driver = await new Builder()
    .forBrowser('chrome')
    .usingServer(`http://${process.env.HS_SELENIUM_ADDRESS}/wd/hub`)
    .build()
  try {
    console.log('>>> Opening page...')
    await driver.get('https://taskworld.com/ja/')
    console.log('>>> Checking...')
    const elements = await driver.findElements(
      By.xpath("//*[contains(text(),'タスクワールド')]")
    )
    if (!elements.length) {
      throw new Error('Nope, cannot find element!')
    }
    console.log(
      '>>> OK, found ' + elements.length + ' elements containing タスクワールド'
    )
  } finally {
    await driver.quit()
  }
})()
