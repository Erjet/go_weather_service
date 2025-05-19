-- V3__update_subscribe_template.sql

UPDATE application_settings
SET value = $$
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Confirm your subscription</title>
  <style>
    body { background-color:#f0f2f5; margin:0; padding:20px; }
    table { border-collapse: collapse; }
    td { padding: 8px; text-align: left; font-family: Arial,sans-serif; }
    h1 { font-size: 24px; color:#333333; margin-bottom:20px; }
    p { font-size: 16px; color:#555555; margin-bottom:30px; }
    a { color:#ffffff; text-decoration:none; font-size:16px; border-radius:5px; }
    .button { display:inline-block; padding:15px 30px; background-color:#28a745; color:#ffffff; text-decoration:none; font-size:16px; border-radius:5px; }
    .footer-link { color:#555555; text-decoration:underline; font-size:14px; }
  </style>
</head>
<body style="background-color:#f0f2f5; margin:0; padding:20px;">
  <table width="100%" cellpadding="0" cellspacing="0">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:40px; text-align:center; font-family:Arial,sans-serif;">
              <h1 style="font-size:24px; color:#333333; margin-bottom:20px;">Confirm your subscription</h1>
              <p style="font-size:16px; color:#555555; margin-bottom:30px;">
                Thank you for signing up for weather updates. Please confirm your subscription by clicking the button below:
              </p>
              <a href="{1}" class="button">
                Confirm subscription
              </a>
              <p style="margin-top:30px; font-size:14px; color:#777777;">
                If you did not request this, you can <a href="{2}" class="footer-link">decline here</a>.
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>
$$
WHERE key = 'EmailSubscribeText';


INSERT INTO application_settings (key, value)
SELECT 'EmailSubscribeText', $$
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Confirm your subscription</title>
  <style>
    body { background-color:#f0f2f5; margin:0; padding:20px; }
    table { border-collapse: collapse; }
    td { padding: 8px; text-align: left; font-family: Arial,sans-serif; }
    h1 { font-size: 24px; color:#333333; margin-bottom:20px; }
    p { font-size: 16px; color:#555555; margin-bottom:30px; }
    a { color:#ffffff; text-decoration:none; font-size:16px; border-radius:5px; }
    .button { display:inline-block; padding:15px 30px; background-color:#28a745; color:#ffffff; text-decoration:none; font-size:16px; border-radius:5px; }
    .footer-link { color:#555555; text-decoration:underline; font-size:14px; }
  </style>
</head>
<body style="background-color:#f0f2f5; margin:0; padding:20px;">
  <table width="100%" cellpadding="0" cellspacing="0">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden;">
          <tr>
            <td style="padding:40px; text-align:center; font-family:Arial,sans-serif;">
              <h1 style="font-size:24px; color:#333333; margin-bottom:20px;">Confirm your subscription</h1>
              <p style="font-size:16px; color:#555555; margin-bottom:30px;">
                Thank you for signing up for weather updates. Please confirm your subscription by clicking the button below:
              </p>
              <a href="{1}" class="button">
                Confirm subscription
              </a>
              <p style="margin-top:30px; font-size:14px; color:#777777;">
                If you did not request this, you can <a href="{2}" class="footer-link">decline here</a>.
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>
$$
WHERE NOT EXISTS (SELECT 1 FROM application_settings WHERE key = 'EmailSubscribeText');
