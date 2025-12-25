我们公司的管理系统需要登陆才能正常使用各种功能，特别是数据查询和监控的功能。

登陆是通过chrome的网页来实现的，可以通过手机扫码来实现，也可以通过在电脑上登陆一个叫宝盒的app来实现，宝盒的安装目录如下：C:\Program Files (x86)\zbox-lite
当宝盒在登陆状态时，打开chrome到公司的管理网址，https://zt-express.com，会自动检测当前的宝盒是否登陆状态。然后会出现一键登录的按钮，用户只需要点击一键登录，即可通过宝盒的认证机制登陆到公司的网络正常使用各种查询和监控管理功能。
	![](assets/zto%20express%20api/file-20251224112348929.png)


宝盒登陆状态不需要持续监控，因为除了每天的第一次登陆的时候浏览器会检测宝盒的状态，并且通过认证取得合适的token，token的有效期如下：
### Token 有效期

|Token|用途|有效期|过期时间|
|---|---|---|---|
|`wyzdzjxhdnh`|长期认证Token|14天|2026-01-06|
|`wyandyy`|短期Session Token|当天有效|每天20:00过期|

示例：
x_ys_dt=c12a5019c5c34ffbb89a30754dad1f3f_GkoDYNYJMDlBQ0FQVQfC3jX8LEEdbWJx; wyzdzjxhdnh=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpYXQiOjE3NjY1NDY2NTgsIm5iZiI6MTc2NjU0NjA1OCwiZXhwIjoxNzY3NzU2MjU4LCJpc3MiOiJjb20uenRvLmNvbm5lY3QiLCJ1dWlkIjoiVHNIdnllTmJYZ2l3M0pYN29OMUg2QSIsImp0aSI6IjIxMDYyNGE1LTUwMTMtNDU2OS04MmY4LTA2N2MxZDUyNzkyZCJ9.BSfE_jnkij5TtnNYICNy7yokqbRyUfYqM35vYXwYv-_2el19IE6umUsGINa-HasyIb5Letz3rq1WzBEqOlJAjpWxOsFNgYBG4961jQ_gkxU72lXEpWFo7j_0K4Ob3QsxfBYyp5cmX8wZpf4f8wWl8tqU7A98HeVePJpHTfGdWmDUiogQFtSPl3S4onmQkH1yqpsyVr2jv_5vIecyr4Nl-VzHFCdCwmuV2O93m3hn-AZdzSrbNWhEDXnXzZksEWIcf-RkU2is5WfYNKukoeWL5UR8PZ_G-bzo2f959h5SSvyrVTV2G9c9pSaXH_UeJbvYiL0Gb0Nd9SSCD0RXGM1Euw; wyandyy=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpYXQiOjE3NjY1NDY2NTgsIm5iZiI6MTc2NjU0NjA1OCwiZXhwIjoxNzY2NjA2NDAwLCJpc3MiOiJjb20uenRvLmNvbm5lY3QiLCJ1dWlkIjoiVHNIdnllTmJYZ2l3M0pYN29OMUg2QSIsInNlc3Npb25pZCI6ImJ5Z1B2MldyVDljLUZMQWpFNVp1RFVVYjczNnl4ay1CM1daTTBlX1hKYXciLCJqdGkiOiI3NmJiYTdjYy1mMWZjLTQxZjMtODJkYS03OGJkODRmNmFlMTQifQ.1T3-_-tUgN_Oi3auTeHGYhBstxKmLOY8YobjUql7rNCT7x058IiI7B_jnDI9C4wP6fqrqBK-0vdMWwSyY_Q5W47h2sffCG9FrDXLqR8GaY_6mvQwE61e5Q6BIURlSxYcGFvrSTr7ctSTb563_POo46RnA8gSGq4Lp7ZvHPDQsQ7w3luv6Qxdchlvnio5VasgT5KdWLBoTOfr7jYlCUZnCiIFB2ASE6ZC55ngUqZ4KDUozvTmdn1ugwikBGUOE7btrd421o_9X3zvgTsCc0jVge0yNuviMZX32wmiYHX8HNmW6ROqNOK33vNW0swc5JTd80yIh7j-jXiJiwBSsBHIYw




只要token在有效时间范围内，哪怕宝盒退出了，网页的登陆状态依然是正常的，可以保持正常的查询功能。应该是已经保存在浏览器的cookies或者持久化配置文件里了。



我现在需要编写一个轻量级的api程序，运行在用户的windows电脑上，代替用户实现宝盒认证这一功能，然后把查询和导出数据等各项操作的请求转化为api的服务，实现用户不需要登陆网页，点开各个功能，才能查询并跟踪各项数据的操作。
部分功能的查询操作我举例如下
1，预约单跟单查询：
curl 'https://preorder-query-center.gw.zt-express.com/preOrderQuery/getSiteOrderTraceList' \
  -H '_catchildmessageid: portal-web-170833621036-490707-0000000332' \
  -H '_catmessageid: portal-web-170833688624-490707-0000000331' \
  -H '_catparentmessageid: portal-web-170833688624-490707-0000000331' \
  -H '_catrootmessageid: portal-web-146034692252-490707-0000000001' \
  -H 'accept: application/json, text/plain, */*' \
  -H 'accept-language: zh-CN,zh;q=0.9,en;q=0.8' \
  -H 'content-type: application/json' \
  -b 'x_ys_dt=c12a5019c5c34ffbb89a30754dad1f3f_GkoDYNYJMDlBQ0FQVQfC3jX8LEEdbWJx; wyzdzjxhdnh=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpYXQiOjE3NjY1NDY2NTgsIm5iZiI6MTc2NjU0NjA1OCwiZXhwIjoxNzY3NzU2MjU4LCJpc3MiOiJjb20uenRvLmNvbm5lY3QiLCJ1dWlkIjoiVHNIdnllTmJYZ2l3M0pYN29OMUg2QSIsImp0aSI6IjIxMDYyNGE1LTUwMTMtNDU2OS04MmY4LTA2N2MxZDUyNzkyZCJ9.BSfE_jnkij5TtnNYICNy7yokqbRyUfYqM35vYXwYv-_2el19IE6umUsGINa-HasyIb5Letz3rq1WzBEqOlJAjpWxOsFNgYBG4961jQ_gkxU72lXEpWFo7j_0K4Ob3QsxfBYyp5cmX8wZpf4f8wWl8tqU7A98HeVePJpHTfGdWmDUiogQFtSPl3S4onmQkH1yqpsyVr2jv_5vIecyr4Nl-VzHFCdCwmuV2O93m3hn-AZdzSrbNWhEDXnXzZksEWIcf-RkU2is5WfYNKukoeWL5UR8PZ_G-bzo2f959h5SSvyrVTV2G9c9pSaXH_UeJbvYiL0Gb0Nd9SSCD0RXGM1Euw; wyandyy=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpYXQiOjE3NjY1NDY2NTgsIm5iZiI6MTc2NjU0NjA1OCwiZXhwIjoxNzY2NjA2NDAwLCJpc3MiOiJjb20uenRvLmNvbm5lY3QiLCJ1dWlkIjoiVHNIdnllTmJYZ2l3M0pYN29OMUg2QSIsInNlc3Npb25pZCI6ImJ5Z1B2MldyVDljLUZMQWpFNVp1RFVVYjczNnl4ay1CM1daTTBlX1hKYXciLCJqdGkiOiI3NmJiYTdjYy1mMWZjLTQxZjMtODJkYS03OGJkODRmNmFlMTQifQ.1T3-_-tUgN_Oi3auTeHGYhBstxKmLOY8YobjUql7rNCT7x058IiI7B_jnDI9C4wP6fqrqBK-0vdMWwSyY_Q5W47h2sffCG9FrDXLqR8GaY_6mvQwE61e5Q6BIURlSxYcGFvrSTr7ctSTb563_POo46RnA8gSGq4Lp7ZvHPDQsQ7w3luv6Qxdchlvnio5VasgT5KdWLBoTOfr7jYlCUZnCiIFB2ASE6ZC55ngUqZ4KDUozvTmdn1ugwikBGUOE7btrd421o_9X3zvgTsCc0jVge0yNuviMZX32wmiYHX8HNmW6ROqNOK33vNW0swc5JTd80yIh7j-jXiJiwBSsBHIYw' \
  -H 'origin: https://www.zt-express.com' \
  -H 'priority: u=1, i' \
  -H 'referer: https://www.zt-express.com/e/order-dispatch-main/newOrderTrace' \
  -H 'sec-ch-ua: "Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Windows"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-site' \
  -H 'user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36' \
  -H 'x-sv-v: pkKbuP0En/KrAkTn39v7JSvzum2pNdHOr6lXbLIA9SsoCPaG+7Pt6VLmAG0TG8zl45fujYx/KGBMwzu1AHPU9ZYbaDx8T8YQfTUY+dH3Z5f0xvNNbvmATDK/52cq2l4q63wTr2Z3cktb9ZMrzT3j/SHzad0oQMVa4f2AV5C3TXW0gSCU6m/RBpbhAysLQpuuZWW2RJ83n2OYbBZO/c+dP9gSRbuYMzhjhnFD0Df1WLyDjNYULca9OozNUjsF/lhZH1b1nIiw4QPAmg7/NZxnAwqg6gQjFD/kVxTp3nA6/aBD8A+MixMb+83TDyA2LdnNZWW2RJ83n2OYbBZO/c+dP9RPHHUJH/STLMxcgrFYOTUFetCgkYZqcYQyu8cnep9d' \
  --data-raw '{"traceQueryChannel":"ALL_PICK_CHANNEL","traceAbnormalMarkQuery":"","traceTimeRequireList":[],"baseList":[],"orderStatusList":[],"searchEmpCodeList":[],"searchSiteCodeList":[],"orderTypeList":[],"partnerIds":[],"traceQueryTime":"ORDER_CREATE_TIME","querySendAddress":"","queryReceiveAddress":"","pickUpCodeStatus":"","payStatus":"","appealStatusList":[],"orderType":0,"startTime":"2025-12-23 00:00:00","endTime":"2025-12-24 23:59:59","querySendProv":"","querySendCity":"","querySendCounty":"","queryReceiveProv":"","queryReceiveCity":"","queryReceiveCounty":"","sortField":"","sortType":0,"pageNum":1,"pageSize":50,"pageIndex":1}'
![](assets/zto%20express%20api/file-20251224112959324.png)

2：订单待办事项查询：
curl 'https://preorder-query-center.gw.zt-express.com/preOrderQuery/getSiteCardFilterCount' \
  -H '_catchildmessageid: portal-web-174133380620-490707-0000000338' \
  -H '_catmessageid: portal-web-174133323228-490707-0000000337' \
  -H '_catparentmessageid: portal-web-174133323228-490707-0000000337' \
  -H '_catrootmessageid: portal-web-146034692252-490707-0000000001' \
  -H 'accept: application/json, text/plain, */*' \
  -H 'accept-language: zh-CN,zh;q=0.9,en;q=0.8' \
  -H 'content-type: application/json' \
  -b 'x_ys_dt=c12a5019c5c34ffbb89a30754dad1f3f_GkoDYNYJMDlBQ0FQVQfC3jX8LEEdbWJx; wyzdzjxhdnh=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpYXQiOjE3NjY1NDY2NTgsIm5iZiI6MTc2NjU0NjA1OCwiZXhwIjoxNzY3NzU2MjU4LCJpc3MiOiJjb20uenRvLmNvbm5lY3QiLCJ1dWlkIjoiVHNIdnllTmJYZ2l3M0pYN29OMUg2QSIsImp0aSI6IjIxMDYyNGE1LTUwMTMtNDU2OS04MmY4LTA2N2MxZDUyNzkyZCJ9.BSfE_jnkij5TtnNYICNy7yokqbRyUfYqM35vYXwYv-_2el19IE6umUsGINa-HasyIb5Letz3rq1WzBEqOlJAjpWxOsFNgYBG4961jQ_gkxU72lXEpWFo7j_0K4Ob3QsxfBYyp5cmX8wZpf4f8wWl8tqU7A98HeVePJpHTfGdWmDUiogQFtSPl3S4onmQkH1yqpsyVr2jv_5vIecyr4Nl-VzHFCdCwmuV2O93m3hn-AZdzSrbNWhEDXnXzZksEWIcf-RkU2is5WfYNKukoeWL5UR8PZ_G-bzo2f959h5SSvyrVTV2G9c9pSaXH_UeJbvYiL0Gb0Nd9SSCD0RXGM1Euw; wyandyy=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpYXQiOjE3NjY1NDY2NTgsIm5iZiI6MTc2NjU0NjA1OCwiZXhwIjoxNzY2NjA2NDAwLCJpc3MiOiJjb20uenRvLmNvbm5lY3QiLCJ1dWlkIjoiVHNIdnllTmJYZ2l3M0pYN29OMUg2QSIsInNlc3Npb25pZCI6ImJ5Z1B2MldyVDljLUZMQWpFNVp1RFVVYjczNnl4ay1CM1daTTBlX1hKYXciLCJqdGkiOiI3NmJiYTdjYy1mMWZjLTQxZjMtODJkYS03OGJkODRmNmFlMTQifQ.1T3-_-tUgN_Oi3auTeHGYhBstxKmLOY8YobjUql7rNCT7x058IiI7B_jnDI9C4wP6fqrqBK-0vdMWwSyY_Q5W47h2sffCG9FrDXLqR8GaY_6mvQwE61e5Q6BIURlSxYcGFvrSTr7ctSTb563_POo46RnA8gSGq4Lp7ZvHPDQsQ7w3luv6Qxdchlvnio5VasgT5KdWLBoTOfr7jYlCUZnCiIFB2ASE6ZC55ngUqZ4KDUozvTmdn1ugwikBGUOE7btrd421o_9X3zvgTsCc0jVge0yNuviMZX32wmiYHX8HNmW6ROqNOK33vNW0swc5JTd80yIh7j-jXiJiwBSsBHIYw' \
  -H 'origin: https://www.zt-express.com' \
  -H 'priority: u=1, i' \
  -H 'referer: https://www.zt-express.com/e/order-dispatch-main/newOrderTrace' \
  -H 'sec-ch-ua: "Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Windows"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-site' \
  -H 'user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36' \
  -H 'x-sv-v: pkKbuP0En/KrAkTn39v7JSvzum2pNdHOr6lXbLIA9SsoCPaG+7Pt6VLmAG0TG8zl45fujYx/KGBMwzu1AHPU9ZYbaDx8T8YQfTUY+dH3Z5f0xvNNbvmATDK/52cq2l4q63wTr2Z3cktb9ZMrzT3j/SHzad0oQMVa4f2AV5C3TXW0gSCU6m/RBpbhAysLQpuuZWW2RJ83n2OYbBZO/c+dP9gSRbuYMzhjhnFD0Df1WLyDjNYULca9OozNUjsF/lhZH1b1nIiw4QPAmg7/NZxnAwqg6gQjFD/kVxTp3nA6/aBZGy0f/zTLDv1CtkE3/5mDZWW2RJ83n2OYbBZO/c+dPgKFMqlwUgDW99QA2NVbn9ThzsrouckaLML0eXV2SqgQ' \
  --data-raw '{"traceQueryChannel":"ALL_PICK_CHANNEL"}'
3 字节省市区数据报表：
curl 'https://orderapi.zt-express.com/opsApi/zjProvinceReport/queryZjPreOrderReport' \
  -H 'Accept-Language: zh-CN,zh;q=0.9,en;q=0.8' \
  -H 'Connection: keep-alive' \
  -b 'x_ys_dt=c12a5019c5c34ffbb89a30754dad1f3f_GkoDYNYJMDlBQ0FQVQfC3jX8LEEdbWJx; wyzdzjxhdnh=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpYXQiOjE3NjY1NDY2NTgsIm5iZiI6MTc2NjU0NjA1OCwiZXhwIjoxNzY3NzU2MjU4LCJpc3MiOiJjb20uenRvLmNvbm5lY3QiLCJ1dWlkIjoiVHNIdnllTmJYZ2l3M0pYN29OMUg2QSIsImp0aSI6IjIxMDYyNGE1LTUwMTMtNDU2OS04MmY4LTA2N2MxZDUyNzkyZCJ9.BSfE_jnkij5TtnNYICNy7yokqbRyUfYqM35vYXwYv-_2el19IE6umUsGINa-HasyIb5Letz3rq1WzBEqOlJAjpWxOsFNgYBG4961jQ_gkxU72lXEpWFo7j_0K4Ob3QsxfBYyp5cmX8wZpf4f8wWl8tqU7A98HeVePJpHTfGdWmDUiogQFtSPl3S4onmQkH1yqpsyVr2jv_5vIecyr4Nl-VzHFCdCwmuV2O93m3hn-AZdzSrbNWhEDXnXzZksEWIcf-RkU2is5WfYNKukoeWL5UR8PZ_G-bzo2f959h5SSvyrVTV2G9c9pSaXH_UeJbvYiL0Gb0Nd9SSCD0RXGM1Euw; wyandyy=eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpYXQiOjE3NjY1NDY2NTgsIm5iZiI6MTc2NjU0NjA1OCwiZXhwIjoxNzY2NjA2NDAwLCJpc3MiOiJjb20uenRvLmNvbm5lY3QiLCJ1dWlkIjoiVHNIdnllTmJYZ2l3M0pYN29OMUg2QSIsInNlc3Npb25pZCI6ImJ5Z1B2MldyVDljLUZMQWpFNVp1RFVVYjczNnl4ay1CM1daTTBlX1hKYXciLCJqdGkiOiI3NmJiYTdjYy1mMWZjLTQxZjMtODJkYS03OGJkODRmNmFlMTQifQ.1T3-_-tUgN_Oi3auTeHGYhBstxKmLOY8YobjUql7rNCT7x058IiI7B_jnDI9C4wP6fqrqBK-0vdMWwSyY_Q5W47h2sffCG9FrDXLqR8GaY_6mvQwE61e5Q6BIURlSxYcGFvrSTr7ctSTb563_POo46RnA8gSGq4Lp7ZvHPDQsQ7w3luv6Qxdchlvnio5VasgT5KdWLBoTOfr7jYlCUZnCiIFB2ASE6ZC55ngUqZ4KDUozvTmdn1ugwikBGUOE7btrd421o_9X3zvgTsCc0jVge0yNuviMZX32wmiYHX8HNmW6ROqNOK33vNW0swc5JTd80yIh7j-jXiJiwBSsBHIYw' \
  -H 'Origin: https://www.zt-express.com' \
  -H 'Referer: https://www.zt-express.com/e/order-operation-new/orderReportForms/byteProv' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Site: same-site' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36' \
  -H '_catChildMessageId: portal-web-175595180824-490707-0000000386' \
  -H '_catMessageId: portal-web-175595137378-490707-0000000385' \
  -H '_catParentMessageId: portal-web-175595137378-490707-0000000385' \
  -H '_catRootMessageId: portal-web-146034692252-490707-0000000001' \
  -H 'accept: application/json, text/plain, */*' \
  -H 'content-type: application/json' \
  -H 'sec-ch-ua: "Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Windows"' \
  --data-raw '{"empCode":"","complianceResult":"","complianceResultQueryList":[],"orderServiceType":[],"startTime":"2025-12-24","batchList":[],"comparisonQueryCode":"","ddzlCompare":"","ddzlCompareType":1,"ddzlCompareCount":null,"provinceName":"","cityName":"","tiktokArea":"","streetName":"","siteCode":"","siteName":"","sortType":1,"sortField":"","endTime":"2025-12-24","whetherPreDepart":null,"whetherNewAreaBatch":null,"queryDateRangeType":1,"pageSize":100,"pageIndex":1}'

有非常多的各种查询数据操作需要用户每天去完成。在网页端进行操作非常的耗时且只能把数据导出为excel之类的零散文件，不方便规模可视化的跟踪处理。

用户需要实现以下目标：
1. 稳定的提供api查询服务，把用户提交的各种查询请求转发给公司的服务器，返回数据。
2. 不占用太多的系统资源 ，以后台服务的形式常驻。
3. 提供简单的日志记录，方便用户查看。
4. 检测登陆状态，当认证失效时，随时调用合适的能力刷新token.
5. 整个程序可移植性高，可以随时换到其它电脑上使用，主要平台都是windows。最好是单个exe文件
6. 有托盘图标可以查看程序的运行状态。
7. 提供简单的测试指令。
8. 对于查询请求有管理队列的能力，不会因为同一时间的请求过多而造成查询失败。
9.  当查询功能失效时，需要程序能自动启动检测宝盒状态，并且重新调用浏览器和宝盒的通讯机制，模拟登陆功能重新获取token，这一情况偶尔会发生，比如用户的宝盒检测到异地登陆的时候，会让所有的浏览器的token失效，需要再次登陆。
10. 提供内网服务，即默认监听0.0.0.0，端口固定为8765
11. 即时响应第三方的查询，效率优先。用户每天会产生大量的高频查询请求。