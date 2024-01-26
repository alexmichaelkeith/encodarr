import os
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))  # noqa
from src.models.system import System  # noqa
from src.models.history import History  # noqa
from src.models.episode import Episode  # noqa
from src.models.season import Season  # noqa
from src.models.series import Series  # noqa
from src.models.setting import Setting  # noqa
from src.models.profile import Profile, profile_codec  # noqa
from src.seeds.seed_system import seed_system  # noqa
from src.seeds.seed_settings import seed_settings  # noqa
from src.seeds.seed_profiles import seed_profiles  # noqa
from sqlalchemy import create_engine, inspect  # noqa
from src.models.base import Base  # noqa


def init_db():

    directory = os.path.dirname("config/db/")
    if not os.path.exists(directory):
        os.makedirs(directory)

    engine = create_engine("sqlite:///config/db/database.db")

    profiles = False
    settings = False
    system = False

    inspector = inspect(engine)
    tables = inspector.get_table_names()
    if 'profiles' not in tables:
        profiles = True
    if 'settings' not in tables:
        settings = True
    if 'system' not in tables:
        system = True

    Base.metadata.create_all(engine)
    conn = engine.connect()
    if profiles:
        seed_profiles(conn)
    if settings:
        seed_settings(conn)
    if system:
        seed_system(conn)


init_db()
