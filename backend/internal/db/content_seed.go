package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Content seeders for Bible passages and the devotional pool.
//
// NOTE ON TEXT SOURCES: to keep the demo free of licensing concerns the
// Indonesian text below is a plain public-domain-style rendering (not the
// copyrighted LAI "Terjemahan Baru"), and the English is the public-domain
// World English Bible (WEB). Swap in a licensed translation for production.

type verse struct {
	n            int
	id, en, mean string
}

type seedChapter struct {
	num    int
	verses []verse
}

type seedBook struct {
	id, name      string // USFM id + Indonesian display name
	ordinal       int    // canonical order 1..66
	totalChapters int    // real total (importer fills the rest)
	testament     string
	chapters      []seedChapter
}

// bibleSeed is a small offline set of beloved passages so the reader works with
// zero setup. `cmd/import-bible` fills in the complete Bible and preserves the
// curated `meaning` notes below (it upserts text only). Book ids are USFM codes,
// matching the importer, so the two never diverge.
var bibleSeed = []seedBook{
	{"PSA", "MAZMUR", 19, 150, "OT", []seedChapter{
		{1, []verse{
			{1, "Berbahagialah orang yang tidak mengikuti nasihat orang fasik, dan tidak berdiri di jalan orang berdosa.", "Blessed is the one who doesn't walk in the counsel of the wicked, nor stand on the path of sinners.", ""},
			{2, "Tetapi yang kesukaannya ialah hukum TUHAN, dan merenungkannya siang dan malam.", "But his delight is in the LORD's law. On his law he meditates day and night.", "Apa yang terus kita renungkan membentuk arah hidup kita — pilihlah dengan sadar."},
			{3, "Ia seperti pohon yang ditanam di tepi aliran air, yang menghasilkan buah pada musimnya, dan daunnya tidak layu.", "He will be like a tree planted by the streams of water, that produces its fruit in its season, whose leaf also does not wither.", ""},
		}},
		{23, []verse{
			{1, "TUHAN adalah gembalaku, aku takkan kekurangan.", "The LORD is my shepherd; I shall lack nothing.", ""},
			{2, "Ia membaringkan aku di padang berumput hijau, Ia menuntun aku ke air yang tenang.", "He makes me lie down in green pastures. He leads me beside still waters.", "\"Padang berumput hijau\" melambangkan penyediaan dan istirahat dari Tuhan — pengingat untuk berhenti sejenak dan percaya bahwa kebutuhanmu dicukupi."},
			{3, "Ia menyegarkan jiwaku, dan menuntun aku di jalan yang benar demi nama-Nya.", "He restores my soul. He guides me in paths of righteousness for his name's sake.", ""},
			{4, "Sekalipun aku berjalan dalam lembah kekelaman, aku tidak takut bahaya, sebab Engkau besertaku; gada dan tongkat-Mu menghiburku.", "Even though I walk through the valley of the shadow of death, I will fear no evil, for you are with me.", "Ketakutan berkurang bukan karena lembahnya hilang, tetapi karena kita tidak melewatinya sendirian."},
			{5, "Engkau menyediakan hidangan bagiku di hadapan lawanku; Engkau mengurapi kepalaku dengan minyak; pialaku penuh melimpah.", "You prepare a table before me in the presence of my enemies. You anoint my head with oil. My cup runs over.", ""},
			{6, "Kebaikan dan kasih setia mengikuti aku seumur hidupku, dan aku akan diam di rumah TUHAN sepanjang masa.", "Surely goodness and loving kindness shall follow me all the days of my life, and I will dwell in the LORD's house forever.", ""},
		}},
	}},
	{"PRO", "AMSAL", 20, 31, "OT", []seedChapter{
		{3, []verse{
			{5, "Percayalah kepada TUHAN dengan segenap hatimu, dan janganlah bersandar pada pengertianmu sendiri.", "Trust in the LORD with all your heart, and don't lean on your own understanding.", "Percaya bukan berarti berhenti berpikir, tetapi tidak menjadikan pengertian sendiri sebagai satu-satunya sandaran."},
			{6, "Akuilah Dia dalam segala lakumu, maka Ia akan meluruskan jalanmu.", "In all your ways acknowledge him, and he will make your paths straight.", ""},
		}},
	}},
	{"JHN", "YOHANES", 43, 21, "NT", []seedChapter{
		{3, []verse{
			{16, "Karena begitu besar kasih Allah akan dunia ini, sehingga Ia mengaruniakan Anak-Nya yang tunggal, supaya setiap orang yang percaya kepada-Nya tidak binasa, melainkan beroleh hidup yang kekal.", "For God so loved the world, that he gave his one and only Son, that whoever believes in him should not perish, but have eternal life.", "Inti kabar baik: kasih yang berinisiatif lebih dulu, sebelum kita layak menerimanya."},
		}},
	}},
	{"PHP", "FILIPI", 50, 4, "NT", []seedChapter{
		{4, []verse{
			{6, "Janganlah khawatir tentang apa pun, tetapi nyatakanlah keinginanmu kepada Allah dalam doa dan permohonan dengan ucapan syukur.", "In nothing be anxious, but in everything, by prayer and petition with thanksgiving, let your requests be made known to God.", "Kekhawatiran diubah menjadi doa; rasa syukur mengingatkan kita pada apa yang sudah ada."},
			{7, "Damai sejahtera Allah yang melampaui segala akal akan memelihara hati dan pikiranmu.", "And the peace of God, which surpasses all understanding, will guard your hearts and your minds.", ""},
			{8, "Akhirnya, pikirkanlah semua yang benar, mulia, adil, suci, manis, dan patut dipuji.", "Finally, whatever things are true, honorable, just, pure, lovely, think about these things.", ""},
			{13, "Aku sanggup menghadapi segala perkara di dalam Dia yang memberi kekuatan kepadaku.", "I can do all things through Christ who strengthens me.", ""},
		}},
	}},
}

func seedBible(ctx context.Context, tx pgx.Tx) error {
	for _, b := range bibleSeed {
		if _, err := tx.Exec(ctx, `
			INSERT INTO bible_books(id,name,ordinal,chapters,testament)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (id) DO NOTHING`,
			b.id, b.name, b.ordinal, b.totalChapters, b.testament); err != nil {
			return err
		}
		for _, ch := range b.chapters {
			for _, v := range ch.verses {
				if _, err := tx.Exec(ctx, `
					INSERT INTO bible_verses(book_id,chapter,verse,text_id,text_en,meaning)
					VALUES ($1,$2,$3,$4,$5,NULLIF($6,''))
					ON CONFLICT (book_id,chapter,verse) DO NOTHING`,
					b.id, ch.num, v.n, v.id, v.en, v.mean); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// devotionalPool is the rotating set of daily devotionals. handleTodayDevotional
// materializes one per calendar day, picked deterministically from the date.
var devotionalPool = []struct {
	title, passage, verse, reflection, prayer string
}{
	{
		"Beristirahat dalam Penyertaan-Nya", "Mazmur 23:1-3",
		"\"Ia membaringkan aku di padang berumput hijau, Ia menuntun aku ke air yang tenang.\"",
		"Di tengah rutinitas yang padat, kita sering lupa berhenti. Hari ini Tuhan mengajak kita percaya bahwa Ia menyediakan yang kita butuhkan — bukan dengan terus berlari, tetapi dengan beristirahat dalam penyertaan-Nya.",
		"Tuhan, tolong aku berhenti sejenak dan percaya pada pemeliharaan-Mu. Amin.",
	},
	{
		"Serahkan Kekhawatiranmu", "Filipi 4:6-7",
		"\"Janganlah khawatir tentang apa pun, tetapi nyatakanlah keinginanmu kepada Allah dalam doa.\"",
		"Kekhawatiran jarang menyelesaikan masalah; ia hanya menguras hari ini. Ubah setiap kecemasan menjadi doa yang jujur, dan biarkan damai sejahtera-Nya menjaga hati dan pikiranmu.",
		"Tuhan, aku serahkan bebanku hari ini kepada-Mu. Berikan aku damai yang melampaui akal. Amin.",
	},
	{
		"Berakar dan Bertumbuh", "Mazmur 1:2-3",
		"\"Ia seperti pohon yang ditanam di tepi aliran air, yang menghasilkan buah pada musimnya.\"",
		"Pertumbuhan sejati tidak instan. Sama seperti pohon yang berakar dekat air, kebiasaan kecil yang setiap hari dirawat akan berbuah pada musimnya. Hari ini, rawat satu akar yang baik.",
		"Tuhan, tanamkan aku dekat sumber-Mu, dan buatlah hidupku berbuah pada waktunya. Amin.",
	},
	{
		"Percaya di Tengah Ketidakpastian", "Amsal 3:5-6",
		"\"Percayalah kepada TUHAN dengan segenap hatimu, dan janganlah bersandar pada pengertianmu sendiri.\"",
		"Ada hal-hal yang tidak bisa kita kendalikan atau mengerti sepenuhnya. Percaya berarti tetap melangkah sambil mengakui Dia — bukan menuntut semua jawaban lebih dulu.",
		"Tuhan, saat aku tidak mengerti, tolong aku tetap percaya dan mengikuti tuntunan-Mu. Amin.",
	},
	{
		"Kasih yang Mendahului", "Yohanes 3:16",
		"\"Karena begitu besar kasih Allah akan dunia ini, sehingga Ia mengaruniakan Anak-Nya yang tunggal.\"",
		"Kasih-Nya tidak menunggu kita sempurna. Ketika kamu merasa kurang atau gagal hari ini, ingat bahwa kamu sudah lebih dulu dikasihi. Dari sanalah kekuatan untuk mencoba lagi bertumbuh.",
		"Tuhan, terima kasih untuk kasih yang tidak kutuntut. Ajar aku hidup dari kasih itu. Amin.",
	},
	{
		"Kekuatan yang Diperbarui", "Filipi 4:13",
		"\"Aku sanggup menghadapi segala perkara di dalam Dia yang memberi kekuatan kepadaku.\"",
		"Kekuatan untuk bertahan tidak selalu datang dari diri sendiri. Ketika tekadmu menipis, mintalah kekuatan yang bukan berasal darimu — dan lanjutkan satu langkah lagi.",
		"Tuhan, saat aku lemah, jadilah kekuatanku hari ini. Amin.",
	},
	{
		"Pikirkan yang Baik", "Filipi 4:8",
		"\"Pikirkanlah semua yang benar, mulia, adil, suci, manis, dan patut dipuji.\"",
		"Apa yang kita izinkan memenuhi pikiran akan membentuk suasana hati dan tindakan. Hari ini, pilih dengan sengaja untuk memikirkan hal-hal yang membangun, bukan yang meruntuhkan.",
		"Tuhan, penuhi pikiranku dengan hal-hal yang baik dan benar. Amin.",
	},
}

func seedDevotionalPool(ctx context.Context, tx pgx.Tx) error {
	for i, d := range devotionalPool {
		if _, err := tx.Exec(ctx, `
			INSERT INTO devotional_pool(id,title,passage,verse_text,reflection,prayer)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (id) DO NOTHING`,
			i, d.title, d.passage, d.verse, d.reflection, d.prayer); err != nil {
			return err
		}
	}
	return nil
}
