import styles from "./MassEditor.module.scss";
import { useContext, useEffect, useRef, useState } from "react";
import { WebSocketContext } from "../../contexts/webSocketContext";
import InputSelect from "../inputs/inputSelect/InputSelect";
import MassEditorToolbar from "../toolbars/massEditorToolbar/MassEditorToolbar";
import sortAndFilter from "../../utils/sortAndFilter";
import MassEditorTable from "../tables/massEditorTable/MassEditorTable";

const MassEditor = () => {
	const wsContext: any = useContext(WebSocketContext);
	const series: any = wsContext?.data?.series;
	const movies: any = wsContext?.data?.movies;
	const settings: any = wsContext?.data?.settings;
	const profiles: any = wsContext?.data?.profiles;
	const [selectedMedia, setSelectedMedia] = useState<any>([]);
	const selectedMediaRef = useRef(selectedMedia);
	const [monitored, setMonitored] = useState<any>(false);
	const [profile, setProfile] = useState<any>();
	const [selected, setSelected] = useState<string | null>(null);
	const [selectAll, setSelectAll] = useState(false);

	const sort = settings?.massEditor_sort;
	const sortDirection = settings?.massEditor_sort_direction;
	const filter = settings?.massEditor_filter;
	const sortedMedia = sortAndFilter(
		series,
		movies,
		profiles,
		sort,
		sortDirection,
		filter
	);

	const handleCheckboxChange = (media: any) => {
		setSelectedMedia((prevSelected: any[]) =>
			prevSelected.some((m) => m.id === media.id)
				? prevSelected.filter((m) => m.id !== media.id)
				: [...prevSelected, media]
		);
	};
	const handleSelectAllChange = () => {
		setSelectAll(!selectAll);
		setSelectedMedia(!selectAll ? sortedMedia : []);
	};

	useEffect(() => {
		selectedMediaRef.current = selectedMedia;
	}, [selectedMedia]);

	useEffect(() => {
		const applyChanges = () => {
			for (const media of selectedMediaRef.current) {
				const type = media?.missing_episodes != undefined ? "series" : "movies";
				media.monitored =
					parseInt(monitored) !== -1 ? parseInt(monitored) : undefined;
				media.profile_id =
					parseInt(profile) !== 0 ? parseInt(profile) : undefined;
				fetch(`/api/${type}/${media.id}`, {
					method: "PUT",
					headers: {
						"Content-Type": "application/json",
						Authorization: `Bearer ${localStorage.getItem("token")}`,
					},

					body: JSON.stringify(media),
				});
			}
		};
		applyChanges();
	}, [monitored, profile]);

	return (
		<div className={styles.massEditor}>
			<MassEditorToolbar
				selected={selected}
				setSelected={setSelected}
				settings={settings}
			/>
			<div className={styles.content}>
				{sortedMedia && sortedMedia.length !== 0 ? (
					<MassEditorTable
						sortedMedia={sortedMedia}
						selectedMedia={selectedMedia}
						selectAll={selectAll}
						handleSelectAllChange={handleSelectAllChange}
						handleCheckboxChange={handleCheckboxChange}
						profiles={profiles}
					/>
				) : (
					<>No Media Found</>
				)}
			</div>
			<div className={styles.footer}>
				<div className={styles.input}>
					<div className={styles.inputContainer}>
						<label className={styles.label}>Monitored </label>
						<InputSelect
							selected={monitored}
							onChange={(e: any) => {
								setMonitored(e.target.value);
							}}
						>
							<option value={-1}>{"No Change"}</option>
							<option value={0}>{"Not Monitored"}</option>
							<option value={1}>{"Monitored"}</option>
						</InputSelect>
					</div>
					<div className={styles.inputContainer}>
						<label className={styles.label}>Profile </label>
						<InputSelect
							selected={profile}
							onChange={(e: any) => {
								setProfile(e.target.value);
							}}
						>
							<option value={0}>{"No Change"}</option>
							{Object.values(profiles || {}).map(
								(profile: any, index: number) => (
									<option value={profile.id} key={index}>
										{profile.name}
									</option>
								)
							)}
						</InputSelect>
					</div>
				</div>
			</div>
		</div>
	);
};
export default MassEditor;
