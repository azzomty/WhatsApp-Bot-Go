import string

def make_map(lower, upper, digits="0123456789"):
    res = {}
    if len(lower) == 26:
        for i, c in enumerate(string.ascii_lowercase): res[c] = lower[i]
    if len(upper) == 26:
        for i, c in enumerate(string.ascii_uppercase): res[c] = upper[i]
    if len(digits) == 10:
        for i, c in enumerate(string.digits): res[c] = digits[i]
    return res

styles = {
    "Math Bold": make_map("𝐚𝐛𝐜𝐝𝐞𝐟𝐠𝐡𝐢𝐣𝐤𝐥𝐦𝐧𝐨𝐩𝐪𝐫𝐬𝐭𝐮𝐯𝐰𝐱𝐲𝐳", "𝐀𝐁𝐂𝐃𝐄𝐅𝐆𝐇𝐈𝐉𝐊𝐋𝐌𝐍𝐎𝐏𝐐𝐑𝐒𝐓𝐔𝐕𝐖𝐗𝐘𝐙", "𝟎𝟏𝟐𝟑𝟒𝟓𝟔𝟕𝟖𝟗"),
    "Math Italic": make_map("𝑎𝑏𝑐𝑑𝑒𝑓𝑔ℎ𝑖𝑗𝑘𝑙𝑚𝑛𝑜𝑝𝑞𝑟𝑠𝑡𝑢𝑣𝑤𝑥𝑦𝑧", "𝐴𝐵𝐶𝐷𝐸𝐹𝐺𝐻𝐼𝐽𝐾𝐿𝑀𝑁𝑂𝑃𝑄𝑅𝑆𝑇𝑈𝑉𝑊𝑋𝑌𝑍"),
    "Math Bold Italic": make_map("𝒂𝒃𝒄𝒅𝒆𝒇𝒈𝒉𝒊𝒋𝒌𝒍𝒎𝒏𝒐𝒑𝒒𝒓𝒔𝒕𝒖𝒗𝒘𝒙𝒚𝒛", "𝑨𝑩𝑪𝑫𝑬𝑭𝑮𝑯𝑰𝑱𝑲𝑳𝑴𝑵𝑶𝑷𝑸𝑹𝑺𝑻𝑼𝑽𝑾𝑿𝒀𝑴"),
    "Fraktur": make_map("𝔞𝔟𝔠𝔡𝔢𝔣𝔤𝔥𝔦𝔧𝔨𝔩𝔪𝔫𝔬𝔭𝔮𝔯𝔰𝔱𝔲𝔳𝔴𝔵𝔶𝔷", "𝔄𝔅ℭ𝔇𝔈𝔉𝔊ℌℑ𝔍𝔎𝔏𝔐𝔑𝔒𝔓𝔔ℜ𝔖𝔗𝔘𝔙𝔚𝔛𝔜ℨ"),
    "Fraktur Bold": make_map("𝖆𝖇𝖈𝖉𝖊𝖋𝖌𝖍𝖎𝖏𝖐𝖑𝖒𝖓𝖔𝖕𝖖𝖗𝖘𝖙𝖚𝖛𝖜𝖝𝖞𝖟", "𝕬𝕭𝕮𝕯𝕰𝕱𝕲𝕳𝕴𝕵𝕶𝕷𝕸𝕹𝕺𝕻𝕼𝕽𝕾𝕿𝖀𝖁𝖂𝖃𝖄𝖅"),
    "Script": make_map("𝒶𝒷𝒸𝒹ℯ𝒻ℊ𝒽𝒾𝒿𝓀𝓁𝓂𝓃ℴ𝓅𝓆𝓇𝓈𝓉𝓊𝓋𝓌𝓍𝓎𝓏", "𝒜ℬ𝒞𝒟ℰℱ𝒢ℋℐ𝒥𝒦ℒℳ𝒩𝒪𝒫𝒬ℛ𝒮𝒯𝒰𝒱𝒲𝒳𝒴𝒵"),
    "Script Bold": make_map("𝓪𝓫𝓬𝓭𝓮𝓯𝓰𝓱𝓲𝓳𝓴𝓵𝓶𝓷𝓸𝓹𝓺𝓻𝓼𝓽𝓾𝓿𝔀𝔁𝔂𝔃", "𝓐𝓑𝓒𝓓𝓔𝓕𝓖𝓗𝓘𝓙𝓚𝓛𝓜𝓝𝓞𝓟𝓠𝓡𝓢𝓣𝓤𝓥𝓦𝓧𝓨𝓩"),
    "Double Struck": make_map("𝕒𝕓𝕔𝕕𝕖𝕗𝕘𝕙𝕚𝕛𝕜𝕝𝕞𝕟𝕠𝕡𝕢𝕣𝕤𝕥𝕦𝕧𝕨𝕩𝕪𝕫", "𝔸𝔹ℂ𝔻𝔼𝔽𝔾ℍ𝕀𝕁𝕂𝕃𝕄ℕ𝕆ℙℚℝ𝕊𝕋𝕌𝕍𝕎𝕏𝕐ℤ", "𝟘𝟙𝟚𝟛𝟜𝟝𝟟𝟠𝟡"),
    "Sans Serif": make_map("𝖺𝖻𝖼𝖽𝖾𝖿𝗀𝗁𝗂𝗃𝗄𝗅𝗆𝗇𝗈𝗉𝗊𝗋𝗌𝗍𝗎𝗏𝗐𝗑𝗒𝗓", "𝖠𝖡𝖢𝖣𝖤𝖥𝖦𝖧𝖨𝖩𝖪𝖫𝖬𝖭𝖮𝖯𝖰𝖱𝖲𝖳𝖴𝖵𝖶𝖷𝖸𝖹", "𝟢𝟣𝟤𝟥𝟦𝟧𝟨𝟩𝟪𝟫"),
    "Sans Serif Bold": make_map("𝗮𝗯𝗰𝗱𝗲𝗳𝗴𝗵𝗶𝗷𝗸𝗹𝗺𝗻𝗼𝗽𝗾𝗿𝘀𝘁𝘂𝘃𝘄𝘅𝘆𝘇", "𝗔𝗕𝗖𝗗𝗘𝗙𝗚𝗛𝗜𝗝𝗞𝗟𝗠𝗡𝗢𝗣𝗤𝗥𝗦𝗧𝗨𝗩𝗪𝗫𝗬𝗭", "𝟬𝟭𝟮𝟯𝟰𝟱𝟲𝟳𝟴𝟵"),
    "Sans Serif Italic": make_map("𝘢𝘣𝘤𝘥𝘦𝘧𝘨𝘩𝘪𝘫𝘬𝘭𝘮𝘯𝘰𝘱𝘲𝘳𝘴𝘵𝘶𝘷𝘸𝘹𝘺𝘻", "𝘈𝘉𝘊𝘋𝘌𝘍𝘎𝘏𝘐𝘑𝘒𝘓𝘔𝘕𝘖𝘗𝘘𝘙𝘚𝘛𝘜𝘝𝘞𝘟𝘠𝘡"),
    "Sans Serif Bold Italic": make_map("𝙖𝙗𝙘𝙙𝙚𝙛𝙜𝙝𝙞𝙟𝙠𝙡𝙢𝙣𝙤𝙥𝙦𝙧𝙨𝙩𝙪𝙫𝙬𝙭𝙮𝙯", "𝘼𝘽𝘾𝘿𝙀𝙁𝙂𝙃𝙄𝙅𝙆𝙇𝙈𝙉𝙊𝙋𝙌𝙍𝙎𝙏𝙐𝙑𝙒𝙓𝙔𝙕"),
    "Monospace": make_map("𝚊𝚋𝚌𝚍𝚎𝚏𝚐𝚑𝚒𝚓𝚔𝚕𝚖𝚗𝚘𝚙𝚚𝚛𝚜𝚝𝚞𝚟𝚠𝚡𝚢𝚣", "𝙰𝙱𝙲𝙳𝙴𝙵𝙶𝙷𝙸𝙹𝙺𝙻𝙼𝙽𝙾𝙿𝚀𝚁𝚂𝚃𝚄𝚅𝚆𝚇𝚈𝚉", "𝟶𝟷𝟸𝟹𝟺𝟻𝟼𝟽𝟾𝟿"),
    "Circled": make_map("ⓐⓑⓒⓓⓔⓕⓖⓗⓘⓙⓚⓛⓜⓝⓞⓟⓠⓡⓢⓣⓤⓥⓦⓧⓨⓩ", "ⒶⒷⒸⒹⒺⒻⒼⒽⒾⒿⓀⓁⓂⓃⓄⓅⓆⓇⓈⓉⓊⓋⓌⓍⓎⓏ", "⓪①②③④⑤⑥⑦⑧⑨"),
    "Circled Negative": make_map("🅐🅑🅒🅓🅔🅕🅖🅗🅘🅙🅚🅛🅜🅝🅞🅟🅠🅡🅢🅣🅤🅥🅦🅧🅨🅩", "🅐🅑🅒🅓🅔🅕🅖🅗🅘🅙🅚🅛🅜🅝🅞🅟🅠🅡🅢🅣🅤🅥🅦🅧🅨🅩", "⓿❶❷❸❹❺❻❼❽❾"),
    "Squared": make_map("🄰🄱🄲🄳🄴🄵🄶🄷🄸🄹🄺🄻🄼🄽🄾🄿🅀🅁🅂🅃🅄🅅🅆🅇🅈🅉", "🄰🄱🄲🄳🄴🄵🄶🄷🄸🄹🄺🄻🄼🄽🄾🄿🅀🅁🅂🅃🅄🅅🅆🅇🅈🅉"),
    "Squared Negative": make_map("🅰🅱🅲🅳🅴🅵🅶🅷🅸🅹🅺🅻🅼🅽🅾🅿🆀🆁🆂🆃🆄🆅🆆🆇🆈🆉", "🅰🅱🅲🅳🅴🅵🅶🅷🅸🅹🅺🅻🅼🅽🅾🅿🆀🆁🆂🆃🆄🆅🆆🆇🆈🆉"),
    "Fullwidth": make_map("ａｂｃｄｅｆｇｈｉｊｋｌｍｎｏｐｑｒｓｔｕｖｗｘｙｚ", "ＡＢＣＤＥＦＧＨＩＪＫＬＭＮＯＰＱＲＳＴＵＶＷＸＹＺ", "０１２３４５６７８９"),
    "Small Caps": make_map("ᴀʙᴄᴅᴇғɢʜɪᴊᴋʟᴍɴᴏᴘǫʀsᴛᴜᴠᴡxʏᴢ", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"),
    "Inverted": make_map("ɐqɔpǝɟƃɥıɾʞlɯuodbɹsʇnʌʍxʎz", "∀ꓭƆᗡƎℲ⅁HIſʞ˥WNOԀΌᴚS⊥∩ΛMX⅄Z"),
    "Reversed": make_map("dɘɔbɒtǫʜiįʞlmnoqrsƚuvwxyz", "AᙠƆᗡƎꟻGHIJK⅃MИOꟼQЯSTUVWXYZ"),
    "Parenthesized": make_map("⒜⒝⒞⒟⒠⒡⒢⒣⒤⒥⒦⒧⒨⒩⒪⒫⒬⒭⒮⒯⒰⒱⒲⒳⒴⒵", "⒜⒝⒞⒟⒠⒡⒢⒣⒤⒥⒦⒧⒨⒩⒪⒫⒬⒭⒮⒯⒰⒱⒲⒳⒴⒵", "⑴⑵⑶⑷⑸⑹⑺⑻⑼"),
    "Cyrillicish": make_map("авсdеfgніјкlмпорqгѕтцvшхуz", "АВСDЕFGНІЈКLМПОРQГЅТЦVШХУZ"),
    "Subscript": make_map("ₐbcdₑfgₕᵢⱼₖₗₘₙₒₚqᵣₛₜᵤᵥwₓyz", "ₐBCDₑFGₕᵢⱼₖₗₘₙₒₚQᵣₛₜᵤᵥWₓYZ", "₀₁₂₃₄₅₆₇₈₉"),
    "Superscript": make_map("ᵃᵇᶜᵈᵉᶠᵍʰⁱʲᵏˡᵐⁿᵒᵖᵠʳˢᵗᵘᵛʷˣʸᶻ", "ᴬᴮᶜᴰᴱᶠᴳᴴᴵᴶᴷᴸᴹᴺᴼᴾQᴿˢᵀᵁⱽᵂˣʸᶻ", "⁰¹²³⁴⁵⁶⁷⁸⁹"),
    "Blue": make_map("🇦🇧🇨🇩🇪🇫🇬🇭🇮🇯🇰🇱🇲🇳🇴🇵🇶🇷🇸🇹🇺🇻🇼🇽🇾🇿", "🇦🇧🇨🇩🇪🇫🇬🇭🇮🇯🇰🇱🇲🇳🇴🇵🇶🇷🇸🇹🇺🇻🇼🇽🇾🇿"),
    "Tiny": make_map("ᵃᵇᶜᵈᵉᶠᵍʰⁱʲᵏˡᵐⁿᵒᵖᵠʳˢᵗᵘᵛʷˣʸᶻ", "ᴬᴮᶜᴰᴱᶠᴳᴴᴵᴶᴷᴸᴹᴺᴼᴾQᴿˢᵀᵁⱽᵂˣʸᶻ"),
    "Comic": make_map("ค๒ς๔єŦﻮђเןкɭ๓ภ๏קợгรՇยשฬץאz", "ค๒ς๔єŦﻮђเןкɭ๓ภ๏קợгรՇยשฬץאZ"),
    "Ancient": make_map("ልጌርዕቿቻኗዘጎጋጕረጠክዐየዒዪነፕሁህሠሸሃፚ", "ልጌርዕቿቻኗዘጎጋጕረጠክዐየዒዪነፕሁህሠሸሃፚ"),
    "Sorcerer": make_map("ǟɮƈɖɛʄɢɦɨʝӄʟʍռօքզʀֆȶʊʋաӼʏʐ", "ǟɮƈɖɛʄɢɦɨʝӄʟʍռօքզʀֆȶʊʋաӼʏʐ"),
    "Special": make_map("αв¢∂єfgнιנкℓмησρqяѕтυνωχуz", "αв¢∂єfgнιנкℓмησρqяѕтυνωχуz"),
    "Weird": make_map("ค๒ƈ๔єŦﻮђเןкɭ๓ภ๏קqгรՇยשฬאץչ", "ค๒ƈ๔єŦﻮђเןкɭ๓ภ๏קqгรՇยשฬאץչ"),
    "Alien": make_map("⏃⏚☊⎅⟒⎎☌⊑⟟⟊☍⌰⋔⋏⍜⌿⍾⍀⌇⏁⎍⎐⍙⌖⊬⋉", "⏃⏚☊⎅⟒⎎☌⊑⟟⟊☍⌰⋔⋏⍜⌿⍾⍀⌇⏁⎍⎐⍙⌖⊬⋉"),
    "Demonic": make_map("a̷b̷c̷d̷e̷f̷g̷h̷i̷j̷k̷l̷m̷n̷o̷p̷q̷r̷s̷t̷u̷v̷w̷x̷y̷z̷", "A̷B̷C̷D̷E̷F̷G̷H̷I̷J̷K̷L̷M̷N̷O̷P̷Q̷R̷S̷T̷U̷V̷W̷X̷Y̷Z̷"),
    "Strike": make_map("a̶b̶c̶d̶e̶f̶g̶h̶i̶j̶k̶l̶m̶n̶o̶p̶q̶r̶s̶t̶u̶v̶w̶x̶y̶z̶", "A̶B̶C̶D̶E̶F̶G̶H̶I̶J̶K̶L̶M̶N̶O̶P̶Q̶R̶S̶T̶U̶V̶W̶X̶Y̶Z̶"),
    "Underline": make_map("a̲b̲c̲d̲e̲f̲g̲h̲i̲j̲k̲l̲m̲n̲o̲p̲q̲r̲s̲t̲u̲v̲w̲x̲y̲z̲", "A̲B̲C̲D̲E̲F̲G̲H̲I̲J̲K̲L̲M̲N̲O̲P̲Q̲R̲S̲T̲U̲V̲W̲X̲Y̲Z̲"),
    "Double Underline": make_map("a̳b̳c̳d̳e̳f̳g̳h̳i̳j̳k̳l̳m̳n̳o̳p̳q̳r̳s̳t̳u̳v̳w̳x̳y̳z̳", "A̳B̳C̳D̳E̳F̳G̳H̳I̳J̳K̳L̳M̳N̳O̳P̳Q̳R̳S̳T̳U̳V̳W̳X̳Y̳Z̳"),
    "Overline": make_map("a̅b̅c̅d̅e̅f̅g̅h̅i̅j̅k̅l̅m̅n̅o̅p̅q̅r̅s̅t̅u̅v̅w̅x̅y̅z̅", "A̅B̅C̅D̅E̅F̅G̅H̅I̅J̅K̅L̅M̅N̅O̅P̅Q̅R̅S̅T̅U̅V̅W̅X̅Y̅Z̅"),
}

# Add combining marks on top of ALL existing styles to make 100+ styles!
diacritics = [
    (" + Slash", "\u0337"),
    (" + Dots", "\u0324"),
    (" + Plus Below", "\u031F"),
    (" + Strikethrough", "\u0336"),
    (" + Wavy", "\u0330"),
    (" + Horns", "\u031B")
]

base_styles = list(styles.items())
for d_name, d_mark in diacritics:
    for base_name, base_map in base_styles[:10]: # only apply to top 10 bases
        new_name = base_name + d_name
        new_map = {k: v + d_mark for k, v in base_map.items()}
        styles[new_name] = new_map

arabic_styles = {
    "Arabic Fatha": "\u064E",
    "Arabic Damma": "\u064F",
    "Arabic Kasra": "\u0650",
    "Arabic Shadda": "\u0651",
    "Arabic Sukun": "\u0652",
    "Arabic Tanween Fath": "\u064B",
    "Arabic Tanween Damm": "\u064C",
    "Arabic Tanween Kasr": "\u064D",
    "Arabic Tatweel": "\u0640",
    "Arabic Maddah": "\u0653",
}

for name, mark in arabic_styles.items():
    res = {}
    arabic_chars = "ابتثجحخدذرزسشصضطظعغفقكلمنهويأإآؤئ"
    for c in arabic_chars:
        if name == "Arabic Tatweel":
            res[c] = c + mark
        else:
            res[c] = c + mark
    styles[name] = res

# Zalgo Generator (we will write a small zalgo string builder in Go, but here we just make 1 mapping)
styles["Zalgo Light"] = {c: c + "\u030D\u0324" for c in string.ascii_lowercase + string.ascii_uppercase}

import io
out = io.StringIO()
out.write("package commands\n\n")
out.write("import \"strings\"\n\n")
out.write("type FontStyle struct {\n")
out.write("\tName string\n")
out.write("\tMapping map[rune]string\n")
out.write("}\n\n")
out.write("var FontStyles = []FontStyle{\n")

for name, mapping in styles.items():
    out.write(f"\t{{\"{name}\", map[rune]string{{\n")
    for k, v in mapping.items():
        out.write(f"\t\t'{k}': \"{v}\",\n")
    out.write("\t}},\n")

out.write("}\n")
out.write("""
func ApplyFont(text string, fontIdx int) string {
    if fontIdx < 0 || fontIdx >= len(FontStyles) {
        return text
    }
    mapping := FontStyles[fontIdx].Mapping
    var sb strings.Builder
    for _, r := range text {
        if val, ok := mapping[r]; ok {
            sb.WriteString(val)
        } else {
            sb.WriteRune(r)
        }
    }
    return sb.String()
}
""")

with open("/home/lennox/Desktop/اهها/Go_Bot/internal/commands/fonts.go", "w") as f:
    f.write(out.getvalue())
