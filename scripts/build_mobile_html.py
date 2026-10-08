import re
import html
import os

def markdown_to_html(text):
    # Convert bold **text**
    text = re.sub(r'\*\*(.+?)\*\*', r'<strong>\1</strong>', text)
    # Convert italic *text*
    text = re.sub(r'\*(.+?)\*', r'<em>\1</em>', text)
    # Convert inline code `code`
    text = re.sub(r'`(.+?)`', r'<code>\1</code>', text)
    # Convert math $math$
    text = re.sub(r'\$(.+?)\$', r'<span class="math">\1</span>', text)
    # Highlight action cues like [Chỉ tay...], [Dừng 1 giây...], [Nhấn giọng...], [Chuyển slide...]
    text = re.sub(r'\[(Chỉ tay[^\]]*|Dừng[^\]]*|Nhấn[^\]]*|Chuyển[^\]]*|Bấm[^\]]*)\]', 
                  r'<span class="stage-cue">🎬 \1</span>', text)
    return text

def parse_markdown():
    with open('docs/kich_ban_thuyet_trinh_tung_slide.md', 'r', encoding='utf-8') as f:
        content = f.read()

    slides = []
    
    # Extract Slide 1 to 13
    slide_pattern = r'## 📌 (SLIDE \d+:[^\n]+)(.*?)(?=\n## 📌|\n# 💻|$)'
    matches = re.findall(slide_pattern, content, re.DOTALL)
    
    for title, body in matches:
        # Extract Image
        img_match = re.search(r'\* \*\*Hình ảnh trên slide:\*\* (.*?)(?=\n\* \*\*|\n>|\n##|$)', body, re.DOTALL)
        image_info = img_match.group(1).strip() if img_match else ""
        
        # Extract Slide Content
        content_match = re.search(r'\* \*\*Nội dung hiển thị trên slide:\*\*(.*?)(?=\n> 🎙️|\n##|$)', body, re.DOTALL)
        slide_bullets = content_match.group(1).strip() if content_match else ""
        
        # Extract Spoken Script
        spoken_match = re.search(r'> 🎙️ \*\*LỜI THOẠI THUYẾT TRÌNH:\*\*(.*?)(?=\n---\n|$)', body, re.DOTALL)
        if not spoken_match:
            spoken_match = re.search(r'> 🎙️ \*\*LỜI THOẠI THUYẾT TRÌNH:\*\*(.*)', body, re.DOTALL)
        spoken_raw = spoken_match.group(1).strip() if spoken_match else ""
        # Clean blockquote '>' markers
        spoken_clean = re.sub(r'^\s*>\s*', '', spoken_raw, flags=re.MULTILINE)
        
        slides.append({
            'type': 'slide',
            'title': title.strip(),
            'image': markdown_to_html(image_info),
            'bullets': markdown_to_html(slide_bullets),
            'spoken': markdown_to_html(spoken_clean)
        })

    # Extract Live Demo
    demo_match = re.search(r'# 💻 PHẦN 2: KỊCH BẢN THUYẾT TRÌNH LIVE DEMO TRỰC TIẾP.*?(?=# 🛡️|$)', content, re.DOTALL)
    demo_content = demo_match.group(0) if demo_match else ""
    
    demo_scenes = []
    scene_matches = re.findall(r'### 🎬 (MÀN \d+:[^\n]+)(.*?)(?=\n### 🎬|\n# 🛡️|$)', demo_content, re.DOTALL)
    for s_title, s_body in scene_matches:
        act_match = re.search(r'\* \*\*Thao tác:\*\*(.*?)(?=\n\* 🎙️|\n###|$)', s_body, re.DOTALL)
        action_text = act_match.group(1).strip() if act_match else ""
        
        voice_match = re.search(r'\* 🎙️ \*\*Lời thoại:\*\*(.*?)(?=\n---|\n###|$)', s_body, re.DOTALL)
        voice_raw = voice_match.group(1).strip() if voice_match else ""
        voice_clean = re.sub(r'^\s*>\s*', '', voice_raw, flags=re.MULTILINE)
        
        demo_scenes.append({
            'title': s_title.strip(),
            'action': markdown_to_html(action_text),
            'voice': markdown_to_html(voice_clean)
        })

    return slides, demo_scenes

def build_html():
    slides, demo_scenes = parse_markdown()

    html_template = f'''<!DOCTYPE html>
<html lang="vi">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
  <title>Kịch Bản Thuyết Trình | Vietnamese Chat Search</title>
  <style>
    :root {{
      --bg-dark: #000000;
      --card-bg: #12151e;
      --card-border: #1e293b;
      --accent-cyan: #00e5ff;
      --accent-blue: #38bdf8;
      --accent-green: #34d399;
      --accent-gold: #f59e0b;
      --text-main: #f1f5f9;
      --text-muted: #94a3b8;
      --speech-bg: #0b192c;
      --speech-border: #2563eb;
      --cue-bg: #854d0e;
      --cue-text: #fef08a;
      --font-base: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    }}

    * {{ box-sizing: border-box; margin: 0; padding: 0; }}

    body {{
      background-color: var(--bg-dark);
      color: var(--text-main);
      font-family: var(--font-base);
      line-height: 1.65;
      font-size: 16px;
      padding-bottom: 90px;
      -webkit-font-smoothing: antialiased;
    }}

    /* STICKY TOP NAV */
    .top-bar {{
      position: sticky;
      top: 0;
      z-index: 100;
      background: rgba(18, 21, 30, 0.95);
      backdrop-filter: blur(12px);
      -webkit-backdrop-filter: blur(12px);
      border-bottom: 1px solid var(--card-border);
      padding: 10px 16px;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }}

    .top-title {{
      font-size: 14px;
      font-weight: 700;
      color: var(--accent-cyan);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      max-width: 180px;
    }}

    .controls {{
      display: flex;
      align-items: center;
      gap: 8px;
    }}

    .btn-ctrl {{
      background: #1e293b;
      border: 1px solid #334155;
      color: var(--text-main);
      border-radius: 8px;
      padding: 6px 12px;
      font-size: 13px;
      font-weight: 600;
      cursor: pointer;
      touch-action: manipulation;
    }}

    .btn-ctrl:active {{
      background: #334155;
    }}

    .tab-pills {{
      display: flex;
      gap: 6px;
      padding: 10px 14px;
      background: #090c14;
      overflow-x: auto;
      scrollbar-width: none;
    }}
    .tab-pills::-webkit-scrollbar {{ display: none; }}

    .pill {{
      padding: 6px 14px;
      border-radius: 20px;
      font-size: 13px;
      font-weight: 600;
      white-space: nowrap;
      background: #1e293b;
      color: var(--text-muted);
      text-decoration: none;
    }}
    .pill.active {{
      background: var(--accent-cyan);
      color: #000;
    }}

    /* MAIN CONTAINER */
    .container {{
      max-width: 680px;
      margin: 0 auto;
      padding: 12px;
    }}

    /* CARD STYLING */
    .card {{
      background: var(--card-bg);
      border: 1px solid var(--card-border);
      border-radius: 16px;
      margin-bottom: 20px;
      overflow: hidden;
      box-shadow: 0 4px 20px rgba(0,0,0,0.5);
    }}

    .card-header {{
      background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
      padding: 14px 16px;
      border-bottom: 1px solid var(--card-border);
      display: flex;
      align-items: center;
      justify-content: space-between;
    }}

    .card-badge {{
      background: #0284c7;
      color: #fff;
      font-size: 11px;
      font-weight: 800;
      padding: 3px 8px;
      border-radius: 6px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }}

    .card-badge.demo {{
      background: #10b981;
    }}

    .card-title {{
      font-size: 16px;
      font-weight: 700;
      color: #fff;
      flex: 1;
      margin-left: 10px;
    }}

    .card-body {{
      padding: 16px;
    }}

    /* META INFO BOX */
    .meta-box {{
      background: #090d16;
      border: 1px solid #1e2638;
      border-radius: 10px;
      padding: 12px;
      margin-bottom: 16px;
      font-size: 13.5px;
    }}

    .meta-row {{
      margin-bottom: 8px;
    }}
    .meta-row:last-child {{ margin-bottom: 0; }}

    .meta-label {{
      font-weight: 700;
      color: var(--accent-blue);
      display: block;
      margin-bottom: 4px;
    }}

    .bullets-content {{
      color: #cbd5e1;
      line-height: 1.55;
    }}

    /* SPEECH PROMPTER BOX */
    .speech-box {{
      background: var(--speech-bg);
      border-left: 4px solid var(--accent-cyan);
      border-radius: 0 12px 12px 0;
      padding: 16px;
      margin-top: 10px;
    }}

    .speech-header {{
      display: flex;
      align-items: center;
      gap: 6px;
      color: var(--accent-cyan);
      font-size: 13px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.5px;
      margin-bottom: 10px;
    }}

    .speech-text {{
      font-size: 17px;
      line-height: 1.75;
      color: #ffffff;
      font-weight: 450;
    }}

    .speech-text p {{
      margin-bottom: 12px;
    }}
    .speech-text p:last-child {{
      margin-bottom: 0;
    }}

    /* STAGE CUE PILL */
    .stage-cue {{
      display: inline-block;
      background: var(--cue-bg);
      color: var(--cue-text);
      font-size: 13px;
      font-weight: 700;
      padding: 2px 8px;
      border-radius: 6px;
      margin: 2px 4px;
      border: 1px solid #ca8a04;
      line-height: 1.3;
    }}

    code {{
      background: #1e293b;
      color: #38bdf8;
      padding: 2px 6px;
      border-radius: 4px;
      font-family: monospace;
      font-size: 14px;
    }}

    .math {{
      color: #f472b6;
      font-weight: 600;
      font-family: monospace;
    }}

    /* BOTTOM FLOATING CONTROLS */
    .bottom-nav {{
      position: fixed;
      bottom: 0;
      left: 0;
      right: 0;
      background: rgba(18, 21, 30, 0.95);
      backdrop-filter: blur(12px);
      -webkit-backdrop-filter: blur(12px);
      border-top: 1px solid var(--card-border);
      padding: 10px 16px;
      display: flex;
      align-items: center;
      justify-content: space-around;
      z-index: 99;
    }}

    .nav-btn {{
      background: #1e293b;
      color: #fff;
      border: 1px solid #334155;
      border-radius: 12px;
      padding: 10px 20px;
      font-size: 15px;
      font-weight: 700;
      flex: 1;
      margin: 0 6px;
      text-align: center;
      cursor: pointer;
    }}
    .nav-btn:active {{ background: #0284c7; }}

    /* FONT SIZE TOGGLE */
    .size-lg .speech-text {{ font-size: 20px; line-height: 1.85; }}
    .size-xl .speech-text {{ font-size: 23px; line-height: 1.95; }}
  </style>
</head>
<body id="body">

  <div class="top-bar">
    <div class="top-title" id="page-indicator">Kịch Bản Thuyết Trình</div>
    <div class="controls">
      <button class="btn-ctrl" onclick="changeFontSize(-1)">A-</button>
      <button class="btn-ctrl" onclick="changeFontSize(1)">A+</button>
      <button class="btn-ctrl" onclick="toggleMode()">💡 Đọc</button>
    </div>
  </div>

  <div class="tab-pills">
    <a href="#slide-1" class="pill active">Slide 1-5</a>
    <a href="#slide-6" class="pill">Slide 6-10</a>
    <a href="#slide-11" class="pill">Slide 11-13</a>
    <a href="#demo-section" class="pill">Live Demo</a>
  </div>

  <div class="container">
'''

    # Render Slides
    for idx, s in enumerate(slides, 1):
        bullets_formatted = s['bullets'].replace('\n', '<br>')
        speech_paragraphs = "".join([f"<p>{p.strip()}</p>" for p in s['spoken'].split('\n\n') if p.strip()])
        
        html_template += f'''
    <div class="card" id="slide-{idx}">
      <div class="card-header">
        <span class="card-badge">Slide {idx}</span>
        <div class="card-title">{s['title']}</div>
      </div>
      <div class="card-body">
        <div class="meta-box">
          <div class="meta-row">
            <span class="meta-label">🖼️ Hình ảnh hiển thị:</span>
            <span class="bullets-content">{s['image']}</span>
          </div>
          <div class="meta-row">
            <span class="meta-label">📝 Điểm chính trên slide:</span>
            <div class="bullets-content">{bullets_formatted}</div>
          </div>
        </div>

        <div class="speech-box">
          <div class="speech-header">🎙️ Lời thoại của em (Nói với anh):</div>
          <div class="speech-text">
            {speech_paragraphs}
          </div>
        </div>
      </div>
    </div>
'''

    # Render Live Demo Section
    html_template += '''
    <div class="card" id="demo-section" style="border-color: #10b981;">
      <div class="card-header" style="background: linear-gradient(135deg, #064e3b 0%, #022c22 100%);">
        <span class="card-badge demo">LIVE DEMO</span>
        <div class="card-title">KỊCH BẢN THỰC CHIẾN 5 PHÚT</div>
      </div>
      <div class="card-body">
        <p style="color:#94a3b8; font-size:13.5px; margin-bottom:14px;">
          Chuẩn bị trước: Chạy sẵn <code>chat_server.exe</code> và mở trình duyệt tại <code>http://localhost:8080</code>.
        </p>
'''

    for idx, d in enumerate(demo_scenes, 1):
        demo_voice = "".join([f"<p>{p.strip()}</p>" for p in d['voice'].split('\n\n') if p.strip()])
        html_template += f'''
        <div style="background:#090d16; border:1px solid #1e293b; border-radius:12px; padding:14px; margin-bottom:14px;">
          <div style="color:#34d399; font-weight:700; font-size:15px; margin-bottom:8px;">🎬 {d['title']}</div>
          <div style="font-size:13.5px; color:#cbd5e1; margin-bottom:10px;">
            <strong style="color:#38bdf8;">👉 Thao tác:</strong> {d['action'].replace(chr(10), '<br>')}
          </div>
          <div class="speech-box" style="margin-top:8px; border-left-color:#34d399;">
            <div class="speech-header" style="color:#34d399;">🎙️ Em nói:</div>
            <div class="speech-text">{demo_voice}</div>
          </div>
        </div>
'''

    html_template += '''
      </div>
    </div>
  </div>

  <div class="bottom-nav">
    <button class="nav-btn" onclick="prevSlide()">◀ Slide Trước</button>
    <button class="nav-btn" onclick="nextSlide()" style="background:#0284c7;">Slide Tiếp ▶</button>
  </div>

  <script>
    let currentSlide = 1;
    const totalSlides = 13;
    let fontSizes = ['', 'size-lg', 'size-xl'];
    let fontIndex = 0;

    function changeFontSize(delta) {
      fontIndex = Math.max(0, Math.min(2, fontIndex + delta));
      const body = document.getElementById('body');
      body.classList.remove('size-lg', 'size-xl');
      if (fontSizes[fontIndex]) {
        body.classList.add(fontSizes[fontIndex]);
      }
    }

    function toggleMode() {
      // Toggle simplified prompter view
      const metas = document.querySelectorAll('.meta-box');
      metas.forEach(m => {
        m.style.display = m.style.display === 'none' ? 'block' : 'none';
      });
    }

    function scrollToSlide(num) {
      currentSlide = Math.max(1, Math.min(totalSlides, num));
      const el = document.getElementById('slide-' + currentSlide);
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'start' });
        document.getElementById('page-indicator').innerText = 'Slide ' + currentSlide + ' / ' + totalSlides;
      }
    }

    function nextSlide() {
      scrollToSlide(currentSlide + 1);
    }

    function prevSlide() {
      scrollToSlide(currentSlide - 1);
    }
  </script>
</body>
</html>
'''

    with open('docs/kich_ban_thuyet_trinh_mobile.html', 'w', encoding='utf-8') as f:
        f.write(html_template)
    print("Done generating docs/kich_ban_thuyet_trinh_mobile.html")

if __name__ == '__main__':
    build_html()
