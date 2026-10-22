package names

import "strings"

// Country catalogs provide a local naming pool for cities not yet given a
// city-specific catalog. Keep these lists stable: generated NPC identities are
// reconstructed from the world seed and generator version.
var countryCatalogs = map[string]Catalog{
	"us": catalog("James John William Robert Charles George Edward Frank Joseph Walter Arthur Henry Thomas Albert Samuel Richard", "Miller Wilson Moore Taylor Anderson Thomas Jackson White Harris Martin Thompson Garcia Martinez Robinson Clark Lewis Lee Walker Hall Allen"),
	"cn": catalog("Wei Fang Jian Ming Li Jun Hui Ping Yan Tao Yong Qiang Lei Chen Guo Xiu", "Wang Li Zhang Liu Chen Yang Huang Zhao Wu Zhou Xu Sun Ma Zhu Hu Guo He Gao Lin"),
	"ru": catalog("Ivan Alexei Boris Dmitri Nikolai Sergei Yuri Viktor Mikhail Pavel Andrei Leonid Fyodor Anatoly Grigori", "Ivanov Smirnov Kuznetsov Popov Sokolov Lebedev Kozlov Novikov Morozov Petrov Volkov Solovyov Vasiliev Zaitsev Pavlov Semyonov"),
	"gb": catalog("Arthur George Edward William James Henry Albert Charles Frederick Thomas Alfred Walter Ernest Harold Frank Leonard", "Smith Jones Taylor Williams Brown Davies Evans Wilson Thomas Roberts Johnson Walker Wright Robinson Thompson White"),
	"de": catalog("Hans Klaus Wolfgang Peter Helmut Heinz Werner Günther Karl Dieter Manfred Uwe Jürgen Horst Rolf Friedrich", "Müller Schmidt Schneider Fischer Weber Meyer Wagner Becker Schulz Hoffmann Schäfer Koch Bauer Richter Klein Wolf"),
	"jp": catalog("Akira Hiroshi Takeshi Kenji Makoto Osamu Satoshi Kazuo Masao Noboru Shigeru Taro Yoshio Haruo Isamu Yoshi", "Sato Suzuki Takahashi Tanaka Watanabe Ito Yamamoto Nakamura Kobayashi Kato Yoshida Yamada Sasaki Yamaguchi Matsumoto Inoue"),
	"fr": catalog("Jean Pierre Michel André Jacques Bernard Louis Marcel René Robert Paul Henri Claude François Georges Alain Maurice", "Martin Bernard Thomas Petit Robert Richard Durand Dubois Moreau Laurent Simon Michel Lefebvre Leroy Roux David Bertrand"),
	"in": catalog("Arjun Raj Dev Anand Vijay Ramesh Suresh Mohan Ravi Prakash Sunil Amit Ashok Dinesh Naresh Manoj", "Sharma Patel Singh Kumar Das Verma Gupta Nair Reddy Iyer Rao Mehta Shah Joshi Mishra"),
	"kr": catalog("Min-jun Seong-ho Jung-ho Yong-su Dong-hyun Jae-won Hyun-woo Byung-ho Sang-ho Joon-ho Tae-young Ki-tae Young-soo Chul-soo Man-seok", "Kim Lee Park Choi Jung Kang Cho Yoon Jang Lim Han Shin Seo Kwon Hwang Ahn"),
	"sa": catalog("Abdullah Abdulrahman Khalid Faisal Mohammed Ahmed Omar Hassan Ibrahim Nasser Saad Tariq Mansour Youssef Salem", "Al-Saud Al-Harbi Al-Otaibi Al-Qahtani Al-Ghamdi Al-Zahrani Al-Shammari Al-Mutairi Al-Dossari Al-Anazi Al-Shehri Al-Rashid Al-Malik Al-Faraj Al-Harthy"),
	"ca": catalog("James John William Robert Charles George Edward Frank Joseph Walter Arthur Henry Thomas Albert Samuel Richard", "Miller Wilson Moore Taylor Anderson Thomas Jackson White Harris Martin Thompson Garcia Martinez Robinson Clark Lewis Lee Walker Hall Allen"),
	"br": catalog("João José Antônio Francisco Carlos Paulo Pedro Manoel Luiz Joaquim Benedito Sebastião Miguel Jorge Alberto Geraldo", "Silva Santos Oliveira Souza Rodrigues Ferreira Alves Pereira Lima Gomes Costa Ribeiro Martins Carvalho Almeida Lopes"),
	"it": catalog("Giuseppe Giovanni Antonio Mario Luigi Franco Bruno Carlo Aldo Giorgio Renato Enzo Vittorio Paolo Angelo Roberto", "Rossi Russo Ferrari Esposito Bianchi Romano Colombo Ricci Marino Greco Bruno Gallo Conti De Luca Costa Mancini"),
	"au": catalog("Arthur George Edward William James Henry Albert Charles Frederick Thomas Alfred Walter Ernest Harold Frank Leonard", "Smith Jones Taylor Williams Brown Davies Evans Wilson Thomas Roberts Johnson Walker Wright Robinson Thompson White"),
	"es": catalog("José Antonio Manuel Francisco Juan Miguel Pedro Rafael Luis Carlos Ramón Vicente Ángel Fernando Emilio Enrique", "García Fernández González Rodríguez López Martínez Sánchez Pérez Gómez Martín Jiménez Ruiz Hernández Díaz Moreno"),
	"tr": catalog("Mehmet Mustafa Ahmet Ali Hasan Hüseyin İbrahim İsmail Yusuf Osman Kemal Süleyman Murat Halil Ömer Nihat", "Yılmaz Kaya Demir Şahin Çelik Yıldız Yıldırım Öztürk Aydın Özdemir Arslan Doğan Kılıç Aslan Çetin Kara"),
	"mx": catalog("José Juan Miguel Francisco Antonio Jesús Manuel Pedro Carlos Ramón Luis Rafael Roberto Jorge Alberto Eduardo", "Hernández García Martínez López González Pérez Rodríguez Sánchez Ramírez Cruz Flores Gómez Morales Vázquez Reyes Jiménez"),
	"id": catalog("Slamet Sutrisno Bambang Joko Agus Budi Hadi Eko Suyanto Suparman Ismail Rahman Abdullah Mulyono Hartono", "Suwandi Santoso Wijaya Saputra Hidayat Kurniawan Setiawan Wibowo Suryanto Gunawan Susanto Nugroho Pratama Kusuma"),
	"ae": catalog("Abdullah Khalid Faisal Mohammed Ahmed Omar Hassan Ibrahim Nasser Saad Tariq Mansour Youssef Salem Rashid", "Al-Mansoori Al-Mazrouei Al-Hameli Al-Shamsi Al-Suwaidi Al-Kaabi Al-Maktoum Al-Falasi Al-Muhairi Al-Dhaheri Al-Qubaisi Al-Ameri Al-Ketbi Al-Marri Al-Neyadi"),
	"ch": catalog("Hans Peter Urs Walter Ernst Fritz Markus Rudolf Werner Paul Andreas Martin Jakob Emil Alfred Robert", "Müller Meier Schmid Keller Weber Huber Schneider Fischer Baumann Gerber Frei Brunner Zimmermann Wenger Steiner Roth"),
	"nl": catalog("Jan Willem Johannes Hendrik Cornelis Pieter Gerrit Dirk Jacobus Antonius Bernardus Petrus Klaas Adrianus Wilhelmus Joost", "De-Jong Jansen De-Vries Van-den-Berg Bakker Janssen Visser Smit Meijer De-Boer Mulder De-Groot Bos Vos Peters Hendriks"),
	"sg": catalog("Ah Seng Ah Kow Ah Huat Ah Chye Ah Beng Ah Soon Ah Kian Ah Leong Ah Meng Ah Tong Ah Wah Ah Keong Ah Fook Ah Teck Ah Boon Ah Joo", "Tan Lim Lee Ng Ong Chua Goh Koh Toh Teo Yeo Chan Chew Low Ho Tay"),
	"il": catalog("Avraham Moshe David Yosef Yaakov Yitzhak Chaim Shlomo Michael Daniel Reuven Menachem Efraim Benjamin Aaron", "Cohen Levi Mizrahi Peretz Biton Dahan Abraham Katz Azulay Ohayon Gabai Amar Ben-David Shapira Malka Dayan"),
	"za": catalog("John James William Hendrik Pieter Johannes Jacobus Themba Sipho Thabo Peter David Michael Daniel Joseph Andrew", "Dlamini Nkosi Botha van der Merwe Naidoo Khumalo Ndlovu Smith Mokoena van Wyk Jacobs Ncube Sithole Molefe Mthembu"),
	"ar": catalog("Juan José Carlos Luis Miguel Jorge Roberto Héctor Raúl Eduardo Julio Ricardo Alberto Daniel Mario Pedro Oscar", "González Rodríguez Gómez Fernández López Martínez Díaz Pérez Sánchez Romero Álvarez Torres Ruiz Flores Acosta Benítez"),
}

// District banks use locally recognizable place words. City names are added
// by the geography generator, making every resulting bank city-scoped.
var countryDistrictBanks = map[string][2][]string{
	"us": {{"Old", "North", "South", "Riverside"}, {"Ward", "Quarter", "Market", "Heights", "Exchange"}},
	"cn": {{"Dongcheng", "Xicheng", "Nancheng", "Beicheng"}, {"Fang", "Li", "Jie", "Market", "Quarter"}},
	"ru": {{"Tsentralny", "Severny", "Yuzhny", "Zavodskoy"}, {"Rayon", "Quarter", "Market", "Station", "Riverside"}},
	"gb": {{"Old", "North", "South", "Riverside"}, {"Ward", "Quarter", "Market", "Heights", "Exchange"}},
	"de": {{"Alt", "Nord", "Sud", "West"}, {"Viertel", "Markt", "Kiez", "Hafen", "Bahnhof"}},
	"jp": {{"Nishi", "Higashi", "Kita", "Minami"}, {"Cho", "Machi", "Ku", "Market", "Quarter"}},
	"fr": {{"Vieux", "Haut", "Bas", "Nouveau"}, {"Quartier", "Marche", "Gare", "Rhone", "Passage"}},
	"in": {{"Purana", "Uttar", "Dakshin", "Naya"}, {"Nagar", "Bazar", "Ward", "Bagh", "Quarter"}},
	"kr": {{"Buk", "Nam", "Dong", "Seo"}, {"Gu", "Market", "Ward", "Station", "Quarter"}},
	"sa": {{"Qadim", "Shamali", "Janubi", "Jadid"}, {"Hayy", "Souq", "Ward", "Gate", "Quarter"}},
	"ca": {{"Old", "North", "South", "Riverside"}, {"Ward", "Quarter", "Market", "Heights", "Exchange"}},
	"br": {{"Velho", "Alto", "Baixo", "Novo"}, {"Bairro", "Mercado", "Estacao", "Jardim", "Porto"}},
	"it": {{"Vecchio", "Alto", "Basso", "Nuovo"}, {"Quartiere", "Mercato", "Stazione", "Porto", "Giardino"}},
	"au": {{"Old", "North", "South", "Riverside"}, {"Ward", "Quarter", "Market", "Heights", "Exchange"}},
	"es": {{"Viejo", "Alto", "Bajo", "Nuevo"}, {"Barrio", "Mercado", "Estacion", "Jardin", "Puerto"}},
	"tr": {{"Eski", "Kuzey", "Guney", "Yeni"}, {"Mahalle", "Pazar", "Istasyon", "Liman", "Quarter"}},
	"mx": {{"Viejo", "Alto", "Bajo", "Nuevo"}, {"Barrio", "Mercado", "Estacion", "Jardin", "Puerto"}},
	"id": {{"Lama", "Utara", "Selatan", "Baru"}, {"Kampung", "Pasar", "Stasiun", "Pelabuhan", "Quarter"}},
	"ae": {{"Qadim", "Shamali", "Janubi", "Jadid"}, {"Hayy", "Souq", "Ward", "Gate", "Quarter"}},
	"ch": {{"Alt", "Nord", "Sud", "West"}, {"Viertel", "Markt", "Kiez", "Hafen", "Bahnhof"}},
	"nl": {{"Oud", "Noord", "Zuid", "Nieuw"}, {"Wijk", "Markt", "Haven", "Gracht", "Station"}},
	"sg": {{"Utara", "Selatan", "Timur", "Barat"}, {"Estate", "Market", "Ward", "Harbour", "Quarter"}},
	"il": {{"Old", "North", "South", "New"}, {"Quarter", "Market", "Port", "Station", "Garden"}},
	"za": {{"Oud", "Noord", "Suid", "Nuwe"}, {"Wyk", "Mark", "Hawe", "Stasie", "Quarter"}},
	"ar": {{"Viejo", "Norte", "Sur", "Nuevo"}, {"Barrio", "Mercado", "Estacion", "Puerto", "Jardin"}},
}

func catalog(given, family string) Catalog {
	return Catalog{GivenNames: splitNames(given), FamilyNames: splitNames(family)}
}

func splitNames(names string) []string {
	return strings.Fields(names)
}
