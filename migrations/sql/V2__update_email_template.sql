-- V2__update_email_template.sql

UPDATE application_settings
SET value = $$
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Weather update</title>
  <style>
    body { background-color:#f0f2f5; margin:0; padding:20px; }
    table { border-collapse: collapse; }
    td, th { padding: 8px; border-bottom: 1px solid #eeeeee; text-align: left; font-size: 14px; color: #333333; }
    th { background-color: #f7f7f7; border-bottom: 2px solid #dddddd; font-weight: bold; }
    h1 { font-size: 24px; color: #333333; margin-bottom: 20px; }
    p { font-size: 16px; color: #555555; margin-bottom: 30px; }
    a { color: #999999; text-decoration: underline; font-size: 12px; }
  </style>
</head>
<body style="background-color:#f0f2f5; margin:0; padding:20px;">
  <table width="100%" cellpadding="0" cellspacing="0">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden; font-family:Arial,sans-serif;">
          <tr>
            <td style="padding:40px; text-align:center;">
              <h1 style="font-size:24px; color:#333333; margin-bottom:20px;">Weather update</h1>
              <p style="font-size:16px; color:#555555; margin-bottom:30px;">
                Here is your weather forecast for the next 3 days:
              </p>
              <table width="100%" cellpadding="8" cellspacing="0" style="border-collapse:collapse;">
                <tr style="background-color:#f7f7f7;">
                  <th align="left" style="border-bottom:2px solid #dddddd; font-size:14px; color:#333333;">Date</th>
                  <th align="left" style="border-bottom:2px solid #dddddd; font-size:14px; color:#333333;">Temperature</th>
                  <th align="left" style="border-bottom:2px solid #dddddd; font-size:14px; color:#333333;">Humidity</th>
                  <th align="left" style="border-bottom:2px solid #dddddd; font-size:14px; color:#333333;">Description</th>
                </tr>
                <tr>
                  <td style="border-bottom:1px solid #eeeeee;">{date1}</td>
                  <td style="border-bottom:1px solid #eeeeee;">{temp1}&deg;C</td>
                  <td style="border-bottom:1px solid #eeeeee;">{hum1}%</td>
                  <td style="border-bottom:1px solid #eeeeee;">{desc1}</td>
                </tr>
                <tr>
                  <td style="border-bottom:1px solid #eeeeee;">{date2}</td>
                  <td style="border-bottom:1px solid #eeeeee;">{temp2}&deg;C</td>
                  <td style="border-bottom:1px solid #eeeeee;">{hum2}%</td>
                  <td style="border-bottom:1px solid #eeeeee;">{desc2}</td>
                </tr>
                <tr>
                  <td style="border-bottom:1px solid #eeeeee;">{date3}</td>
                  <td style="border-bottom:1px solid #eeeeee;">{temp3}&deg;C</td>
                  <td style="border-bottom:1px solid #eeeeee;">{hum3}%</td>
                  <td style="border-bottom:1px solid #eeeeee;">{desc3}</td>
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td style="padding:20px; text-align:center; font-size:12px; color:#999999;">
              <a href="{unsubscribe_url}" style="color:#999999; text-decoration:underline;">Unsubscribe</a>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>
$$
WHERE key = 'EmailUpdateText';

INSERT INTO application_settings (key, value)
SELECT 'EmailUpdateText', $$
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Weather update</title>
</head>
<body style="background-color:#f0f2f5; margin:0; padding:20px;">
  <table width="100%" cellpadding="0" cellspacing="0">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden; font-family:Arial,sans-serif;">
          <tr>
            <td style="padding:40px; text-align:center;">
              <h1 style="font-size:24px; color:#333333; margin-bottom:20px;">Weather update</h1>
              <p style="font-size:16px; color:#555555; margin-bottom:30px;">
                Here is your weather forecast for the next 3 days:
              </p>
              <table width="100%" cellpadding="8" cellspacing="0" style="border-collapse:collapse;">
                <tr style="background-color:#f7f7f7;">
                  <th align="left" style="border-bottom:2px solid #dddddd; font-size:14px; color:#333333;">Date</th>
                  <th align="left" style="border-bottom:2px solid #dddddd; font-size:14px; color:#333333;">Temperature</th>
                  <th align="left" style="border-bottom:2px solid #dddddd; font-size:14px; color:#333333;">Humidity</th>
                  <th align="left" style="border-bottom:2px solid #dddddd; font-size:14px; color:#333333;">Description</th>
                </tr>
                <tr>
                  <td style="border-bottom:1px solid #eeeeee;">{date1}</td>
                  <td style="border-bottom:1px solid #eeeeee;">{temp1}&deg;C</td>
                  <td style="border-bottom:1px solid #eeeeee;">{hum1}%</td>
                  <td style="border-bottom:1px solid #eeeeee;">{desc1}</td>
                </tr>
                <tr>
                  <td style="border-bottom:1px solid #eeeeee;">{date2}</td>
                  <td style="border-bottom:1px solid #eeeeee;">{temp2}&deg;C</td>
                  <td style="border-bottom:1px solid #eeeeee;">{hum2}%</td>
                  <td style="border-bottom:1px solid #eeeeee;">{desc2}</td>
                </tr>
                <tr>
                  <td style="border-bottom:1px solid #eeeeee;">{date3}</td>
                  <td style="border-bottom:1px solid #eeeeee;">{temp3}&deg;C</td>
                  <td style="border-bottom:1px solid #eeeeee;">{hum3}%</td>
                  <td style="border-bottom:1px solid #eeeeee;">{desc3}</td>
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td style="padding:20px; text-align:center; font-size:12px; color:#999999;">
              <a href="{unsubscribe_url}" style="color:#999999; text-decoration:underline;">Unsubscribe</a>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>
$$
WHERE NOT EXISTS (SELECT 1 FROM application_settings WHERE key = 'EmailUpdateText');
