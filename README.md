# FlameChat


這是用Cursor AI產生的專案，下面是提問內容


我要做一個聊天室，後端名稱goserve使用golang，前端名稱vueweb使用vue。

顯示聊天房間列表，使用 rest api實作，用網址對應 '/getrooms' , 回傳資料有 roomname , roomid

goserver需要儲存歷史聊天紀錄到chathistory，資料包含暱稱、內容、時間，最多存200筆資料，如果超過200筆資料先移除最舊的資料再加入新的

使用者點選聊天房間列表中任一房間後可進入聊天室，使用websocket進行連線，然後發送訊息 protocol : joinroom , roomid ，goserve確認有房間後回傳 protocol : resjoinroom , status : ok , Chathistory ，前端將切換到聊天頁面，
如果沒有roomid聊天室就關閉websocket,並將chathistory顯示在訊息顯示框

使用者也可以點選開新房間按鈕建立新聊天房間，使用websocket進行連線，然後發送訊息 protocol : openroom , roomname ，goserve新建房間資料後回傳 protocol : resopenroom , status : ok , roomid，前端將切換到聊天頁面，

進入聊天房間後，會使用websocket連接，房間介面可以設定暱稱，有輸入框可以輸入要發送的訊息，按下發送按鈕就會發送message到goserve，goserve收到後會將資料儲存到chathistory，
然後對在這個房間的使用者廣播這一筆資料，發送功能使用websocket發送，
資料封包需要包含protocol、暱稱、訊息、時間。

訊息顯示框，會將自己的訊息放左邊，其他人的訊息放右邊，單筆訊息內容需分成3行，暱稱第一行，內容第2行，時間第3行，我自己的訊息框背景顏色8ad690，其他人的顏色9ed7ff

