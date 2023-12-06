import styles from "./Profile.module.scss";
import AddIcon from "@mui/icons-material/Add";
const Profile = ({ name, type = "", codecs }: any) => {
	return (
		<div className={styles.profile} key={name}>
			{type === "add" ? (
				<div className={styles.add}>
					<div className={styles.box}>{<AddIcon fontSize="large" />}</div>
				</div>
			) : (
				<div className={styles.normal}>
					<div className={styles.header}>{name}</div>
					<div className={styles.codecs}>
						{codecs.map((codec: any) => (
							<div className={styles.codec} key={name + "-" + codec}>
								{codec}
							</div>
						))}
					</div>
				</div>
			)}
		</div>
	);
};
export default Profile;
