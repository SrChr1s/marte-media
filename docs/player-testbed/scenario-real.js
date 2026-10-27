(async()=>{
const sleep=ms=>new Promise(r=>setTimeout(r,ms));
const log=m=>window.runtime.LogError('[real] '+m);
const until=async(fn,ms,label)=>{const t0=Date.now();while(Date.now()-t0<ms){try{if(fn())return true}catch(e){}await sleep(300)}log('TIMEOUT waiting for '+label);return false};
const rail=l=>[...document.querySelectorAll('.rail .item')].find(b=>b.textContent.includes(l)).click();
const search=q=>{const i=document.querySelector('.bar input');i.value=q;i.dispatchEvent(new Event('input',{bubbles:true}));document.querySelector('.bar').requestSubmit()};
const shot=async n=>{log('SHOT '+n); await sleep(2500)};
try{
localStorage.removeItem('marte-media.player');localStorage.setItem('marte-media.debug','1');
await sleep(1500);
rail('Reproducir'); await sleep(500);
const t0=Date.now();
search('marte documental');
await until(()=>document.querySelectorAll('.vcard').length>0,60000,'video results');
log('video search: '+document.querySelectorAll('.vcard').length+' results in '+(Date.now()-t0)+'ms; first="'+(document.querySelector('.vtitle')||{}).textContent+'"');
await shot('r1_results');
const t1=Date.now();
document.querySelector('.vcard .open').click();
const M=window.__marte;
await until(()=>M&&M.player.status==='playing'&&M.video.currentTime>1.5,90000,'video playing');
const st=M.player.stream;
log('video start: '+(Date.now()-t1)+'ms status='+M.player.status+' picked='+(st&&st.video&&st.video.mime)+' h='+(st&&st.video&&st.video.height)+' heights='+JSON.stringify(M.player.heights)+' vsize='+M.video.videoWidth+'x'+M.video.videoHeight);
await sleep(1500);
await shot('r2_watching');
const ds=[];const q0=M.video.getVideoPlaybackQuality();
for(let i=0;i<30;i++){ds.push(M.audio.currentTime-M.video.currentTime);await sleep(500)}
const ab=ds.map(Math.abs),q1=M.video.getVideoPlaybackQuality(),srt=[...ab].sort((a,b)=>a-b);
log('drift(15s) n='+ds.length+' maxAbs='+Math.max(...ab).toFixed(3)+' p90='+srt[Math.floor(srt.length*.9)].toFixed(3)+' meanAbs='+(ab.reduce((a,b)=>a+b,0)/ab.length).toFixed(3)+' audioT='+M.audio.currentTime.toFixed(1)+' dropped='+(q1.droppedVideoFrames-q0.droppedVideoFrames)+'/'+(q1.totalVideoFrames-q0.totalVideoFrames)+' status='+M.player.status);
const sk=M.audio.currentTime+90; M.seek(sk); await sleep(4000);
log('seek +90s: a='+M.audio.currentTime.toFixed(2)+' v='+M.video.currentTime.toFixed(2)+' status='+M.player.status+' drift='+(M.audio.currentTime-M.video.currentTime).toFixed(3));
const a0=M.audio.currentTime; await M.setQuality(M.player.heights[M.player.heights.length-1]); await sleep(5000);
log('video.error='+JSON.stringify(M.video.error&&{c:M.video.error.code,m:M.video.error.message})+' src='+(M.video.currentSrc||'').slice(0,60)+' paused='+M.video.paused+' ready='+M.video.readyState);
log('quality->'+M.player.heights[M.player.heights.length-1]+'p: vsize='+M.video.videoWidth+'x'+M.video.videoHeight+' audioMoved='+(M.audio.currentTime-a0).toFixed(2)+' drift='+(M.audio.currentTime-M.video.currentTime).toFixed(3)+' status='+M.player.status);
document.querySelector('.back').click(); await sleep(600);
[...document.querySelectorAll('.choice button')].find(b=>b.textContent.trim()==='Música').click();
await sleep(500); search('daft punk get lucky');
const t2=Date.now();
await until(()=>document.querySelectorAll('.song').length>0,60000,'music results');
log('music search: '+document.querySelectorAll('.song').length+' songs in '+(Date.now()-t2)+'ms; first="'+(document.querySelector('.song .title')||{}).textContent+'"');
await shot('r3_music');
const t3=Date.now();
document.querySelector('.song').click();
await until(()=>M.player.status==='playing'&&M.audio.currentTime>1.5&&M.player.mode==='audio',90000,'music playing');
const cur=M.player.queue[M.player.index];
log('music start: '+(Date.now()-t3)+'ms mode='+M.player.mode+' hasVideo='+M.player.hasVideo+' title="'+cur.title+'" channel="'+cur.channel+'" dur='+cur.duration+' audio='+(M.player.stream.audio&&M.player.stream.audio.mime));
const am=M.audio.currentTime; await sleep(3000);
log('music audioMoved(3s)='+(M.audio.currentTime-am).toFixed(2));
await shot('r4_music_playing');
log('TRAIL\n'+window.__marte.trail.join('\n'));
}catch(e){log('SCRIPT ERROR '+e+' '+(e&&e.stack||''))}
log('done');
})()
