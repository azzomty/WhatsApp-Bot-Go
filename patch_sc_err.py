with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'r') as f:
    content = f.read()

content = content.replace('sendMessage(ctx, "حدث خطأ أثناء البحث في ساوند كلاود.\\nقد لا توجد نتائج.")', 'sendMessage(ctx, "حدث خطأ أثناء البحث في ساوند كلاود:\\n" + string(out))')

with open('/home/lennox/Desktop/اهها/Go_Bot/internal/commands/downloader.go', 'w') as f:
    f.write(content)
